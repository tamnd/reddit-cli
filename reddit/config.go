package reddit

import (
	"os"
	"path/filepath"
	"time"
)

// Site constants.
const (
	BaseURL    = "https://www.reddit.com"
	OldBaseURL = "https://old.reddit.com"

	DefaultDelay    = 2 * time.Second
	DefaultWorkers  = 2
	DefaultTimeout  = 30 * time.Second
	DefaultRetries  = 3
	DefaultCacheTTL = 24 * time.Hour

	// MaxPageLimit is the most things Reddit returns in one listing page.
	MaxPageLimit = 100
)

// DefaultUserAgent identifies the tool and links back to its source. Reddit
// rate-limits empty and generic browser agents the hardest, and asks clients to
// send a descriptive, unique one, so that is the default. Override it with
// --user-agent or Config.UserAgent.
const DefaultUserAgent = "reddit-cli (+https://github.com/tamnd/reddit-cli)"

// browserAgents back the rare HTML fallback paths, where a page renders
// differently for a browser agent than for a bot one.
var browserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:132.0) Gecko/20100101 Firefox/132.0",
}

// Config holds runtime configuration for the library.
type Config struct {
	DataDir    string        // root for cache/store; defaults to the XDG data dir
	Workers    int           // concurrency for multi-page/bulk work
	Delay      time.Duration // minimum spacing between requests
	Timeout    time.Duration // per-request timeout
	Retries    int           // retry attempts on 429/5xx
	CacheTTL   time.Duration // on-disk page-cache freshness window
	NoCache    bool          // bypass the on-disk page cache entirely
	Refresh    bool          // force re-fetch and overwrite the cache
	UserAgent  string        // override the request User-Agent
	CookiePath string        // optional Netscape cookie jar for a lent session
}

// DefaultConfig returns polite defaults rooted at the XDG data dir.
func DefaultConfig() Config {
	return Config{
		DataDir:   DefaultDataDir(),
		Workers:   DefaultWorkers,
		Delay:     DefaultDelay,
		Timeout:   DefaultTimeout,
		Retries:   DefaultRetries,
		CacheTTL:  DefaultCacheTTL,
		UserAgent: DefaultUserAgent,
	}
}

// DefaultDataDir resolves $XDG_DATA_HOME/reddit (or ~/.local/share/reddit).
func DefaultDataDir() string {
	if d := os.Getenv("REDDIT_DATA_DIR"); d != "" {
		return d
	}
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, "reddit")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), "reddit")
	}
	return filepath.Join(home, ".local", "share", "reddit")
}

// CacheDir is where gzipped page captures live.
func (c Config) CacheDir() string { return filepath.Join(c.DataDir, "cache") }

// StorePath is the default path of the optional SQLite store.
func (c Config) StorePath() string { return filepath.Join(c.DataDir, "reddit.db") }
