package reddit

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func unixTime(ts int64) time.Time { return time.Unix(ts, 0) }

// LoadCookies reads a Netscape/Mozilla cookies.txt file (the format exported by
// most browser extensions and by curl) into http.Cookie values. A lent session
// lifts some of Reddit's anonymous limits and reaches quarantined communities.
func LoadCookies(path string) ([]*http.Cookie, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var cookies []*http.Cookie
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		// "#HttpOnly_" prefixed lines are real cookies; bare "#" lines are comments.
		httpOnly := false
		if strings.HasPrefix(line, "#HttpOnly_") {
			httpOnly = true
			line = strings.TrimPrefix(line, "#HttpOnly_")
		} else if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}
		c := &http.Cookie{
			Domain:   fields[0],
			Path:     fields[2],
			Secure:   strings.EqualFold(fields[3], "TRUE"),
			Name:     fields[5],
			Value:    fields[6],
			HttpOnly: httpOnly,
		}
		if ts, err := strconv.ParseInt(fields[4], 10, 64); err == nil && ts > 0 {
			c.Expires = unixTime(ts)
		}
		cookies = append(cookies, c)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(cookies) == 0 {
		return nil, fmt.Errorf("no cookies found in %s", path)
	}
	return cookies, nil
}
