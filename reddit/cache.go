package reddit

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Cache is a content-addressed, gzipped on-disk response cache. A captured
// response is stored under cache/<ab>/<sha256>.json.gz, keyed by its URL.
type Cache struct {
	dir string
	ttl time.Duration
}

// NewCache returns a cache rooted at cfg.CacheDir() (created on first write).
func NewCache(cfg Config) *Cache {
	ttl := cfg.CacheTTL
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}
	return &Cache{dir: cfg.CacheDir(), ttl: ttl}
}

func (c *Cache) pathFor(url string) string {
	sum := sha256.Sum256([]byte(url))
	h := hex.EncodeToString(sum[:])
	return filepath.Join(c.dir, h[:2], h+".json.gz")
}

// Get returns cached bytes for a URL when present and within the TTL.
func (c *Cache) Get(url string) ([]byte, bool) {
	p := c.pathFor(url)
	info, err := os.Stat(p)
	if err != nil {
		return nil, false
	}
	if c.ttl > 0 && time.Since(info.ModTime()) > c.ttl {
		return nil, false
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, false
	}
	defer func() { _ = f.Close() }()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, false
	}
	defer func() { _ = zr.Close() }()
	body, err := io.ReadAll(zr)
	if err != nil {
		return nil, false
	}
	return body, true
}

// Put writes bytes for a URL into the cache.
func (c *Cache) Put(url string, body []byte) error {
	p := c.pathFor(url)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	tmp := p + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	zw := gzip.NewWriter(f)
	if _, err := zw.Write(body); err != nil {
		_ = zw.Close()
		_ = f.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// Path returns the on-disk cache path for a URL (whether or not it exists).
func (c *Cache) Path(url string) string { return c.pathFor(url) }

// Clear removes the entire cache directory.
func (c *Cache) Clear() error { return os.RemoveAll(c.dir) }

// Stats reports the file count and total bytes held in the cache.
func (c *Cache) Stats() (files int, total int64, err error) {
	walkErr := filepath.Walk(c.dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			files++
			total += info.Size()
		}
		return nil
	})
	if os.IsNotExist(walkErr) {
		return 0, 0, nil
	}
	return files, total, walkErr
}

// cachedFetch returns body bytes for a URL, reading the cache unless it is off
// or a refresh was asked for, and populating it on a miss. It only caches a
// clean 200 response so error envelopes are always re-fetched.
func (c *Client) cachedFetch(ctx context.Context, url string) ([]byte, int, error) {
	if c.cache != nil && !c.cfg.NoCache && !c.cfg.Refresh {
		if body, ok := c.cache.Get(url); ok {
			return body, 200, nil
		}
	}
	body, code, err := c.Fetch(ctx, url)
	if err != nil {
		return body, code, err
	}
	if code == 200 && c.cache != nil && !c.cfg.NoCache {
		_ = c.cache.Put(url, body)
	}
	return body, code, nil
}
