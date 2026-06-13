package reddit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// Errors the library surfaces. Callers map them to exit codes.
var (
	// ErrBlocked signals that Reddit refused the page: an interstitial
	// challenge, a redirect to a block or login page, or a 403 that is not a
	// private community.
	ErrBlocked = errors.New("blocked: reddit refused the request")
	// ErrRateLimited signals a sustained HTTP 429 after retries.
	ErrRateLimited = errors.New("rate limited (HTTP 429)")
	// ErrNotFound is returned when a thing 404s.
	ErrNotFound = errors.New("not found")
	// ErrPrivate is returned for a private community (403 with reason private).
	ErrPrivate = errors.New("private community")
	// ErrBanned is returned for a banned community (404 with reason banned).
	ErrBanned = errors.New("banned community")
)

// Client performs polite, retrying HTTP GETs against reddit.com.
type Client struct {
	http      *http.Client
	userAgent string
	limiter   *rate.Limiter
	retries   int
	cfg       Config
	cache     *Cache
}

// newLimiter builds a token-bucket limiter: rate = workers/delay, burst =
// workers, so every worker can fire at startup but the average pace stays
// polite.
func newLimiter(cfg Config) *rate.Limiter {
	if cfg.Delay <= 0 || cfg.Workers <= 0 {
		return rate.NewLimiter(rate.Inf, 1)
	}
	r := rate.Limit(float64(cfg.Workers) / cfg.Delay.Seconds())
	return rate.NewLimiter(r, cfg.Workers)
}

func transport(cfg Config) *http.Transport {
	return &http.Transport{
		MaxIdleConns:        cfg.Workers + 4,
		MaxConnsPerHost:     cfg.Workers + 4,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
}

// NewClient builds an anonymous client.
func NewClient(cfg Config) *Client {
	retries := cfg.Retries
	if retries <= 0 {
		retries = DefaultRetries
	}
	ua := cfg.UserAgent
	if ua == "" {
		ua = DefaultUserAgent
	}
	return &Client{
		http:      &http.Client{Timeout: cfg.Timeout, Transport: transport(cfg)},
		userAgent: ua,
		limiter:   newLimiter(cfg),
		retries:   retries,
		cfg:       cfg,
	}
}

// WithCache attaches a response cache. Subsequent reads consult it (honoring
// the config's NoCache and Refresh flags) and populate it on a clean miss.
func (c *Client) WithCache(cache *Cache) *Client {
	c.cache = cache
	return c
}

// NewClientWithCookies builds a client pre-loaded with a lent session.
func NewClientWithCookies(cfg Config, cookies []*http.Cookie) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse(BaseURL)
	jar.SetCookies(u, cookies)
	c := NewClient(cfg)
	c.http.Jar = jar
	return c, nil
}

// Fetch returns the raw body and status for a URL, retrying transient failures.
// A 404 returns (body, 404, nil) so callers can read a reason envelope. A
// refusal returns ErrBlocked; a sustained 429 returns ErrRateLimited.
func (c *Client) Fetch(ctx context.Context, rawurl string) ([]byte, int, error) {
	maxAttempts := c.retries
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, 0, err
		}
		body, code, err := c.doGet(ctx, rawurl)
		if err != nil {
			if errors.Is(err, ErrBlocked) || attempt == maxAttempts {
				return nil, code, err
			}
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		switch {
		case code == 429:
			if attempt == maxAttempts {
				return nil, code, ErrRateLimited
			}
			time.Sleep(time.Duration(attempt*attempt) * 5 * time.Second)
			continue
		case code == 403, code == 404:
			// Hand the body back; the caller reads {"reason":...} envelopes.
			return body, code, nil
		case code >= 500:
			if attempt == maxAttempts {
				return nil, code, fmt.Errorf("server error HTTP %d", code)
			}
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
			continue
		}
		return body, code, nil
	}
	return nil, 0, fmt.Errorf("all %d attempts failed", maxAttempts)
}

func (c *Client) doGet(ctx context.Context, rawurl string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, 0, err
	}
	ua := c.userAgent
	// The HTML fallback host renders better for a browser agent.
	if strings.Contains(rawurl, "old.reddit.com") {
		ua = browserAgents[rand.Intn(len(browserAgents))]
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "application/json,text/html;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	// A redirect to a login or block page means refused.
	final := resp.Request.URL.String()
	if isBlockURL(final) {
		return nil, 403, fmt.Errorf("%w (%s redirected to %s)", ErrBlocked, rawurl, final)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	// The "whoa there, pardner!" interstitial is served as HTML with a 200 or a
	// 403; treat it as refused regardless of status.
	if looksBlocked(body) {
		return nil, resp.StatusCode, fmt.Errorf("%w (interstitial challenge for %s)", ErrBlocked, rawurl)
	}
	return body, resp.StatusCode, nil
}

func isBlockURL(u string) bool {
	low := strings.ToLower(u)
	return strings.Contains(low, "/login") ||
		strings.Contains(low, "/account/login") ||
		strings.Contains(low, "blocked") ||
		strings.Contains(low, "/over18")
}

// looksBlocked reports whether an HTML body is one of Reddit's challenge or
// block interstitials. JSON bodies never match, so a normal .json response
// passes through.
func looksBlocked(body []byte) bool {
	if len(body) == 0 {
		return false
	}
	if body[0] == '{' || body[0] == '[' {
		return false
	}
	s := strings.ToLower(string(body[:min(len(body), 4096)]))
	return strings.Contains(s, "whoa there, pardner") ||
		strings.Contains(s, "you've been blocked") ||
		strings.Contains(s, "blocked by network security")
}
