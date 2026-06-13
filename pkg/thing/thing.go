// Package thing models Reddit's "thing" taxonomy: the kind prefixes (t1..t6),
// the base-36 ids, and the fullnames that join them (kind_id, e.g. t3_abc123).
// It also classifies any Reddit URL or bare id into a (kind, id) pair. It has no
// knowledge of HTTP or the rest of reddit-cli, so it is safe to import anywhere.
package thing

import "strings"

// Kind is a Reddit thing prefix.
type Kind string

// The kinds Reddit assigns. t4 (message) is unreachable without a signed-in
// session and t6 (award) only shows up as counts, but both are modeled for
// completeness.
const (
	KindComment   Kind = "t1"
	KindAccount   Kind = "t2"
	KindLink      Kind = "t3"
	KindMessage   Kind = "t4"
	KindSubreddit Kind = "t5"
	KindAward     Kind = "t6"
)

// Name maps a kind to the word reddit-cli uses for it in records and the id
// command: post, comment, subreddit, user. Unknown kinds return "".
func (k Kind) Name() string {
	switch k {
	case KindLink:
		return "post"
	case KindComment:
		return "comment"
	case KindSubreddit:
		return "subreddit"
	case KindAccount:
		return "user"
	case KindMessage:
		return "message"
	case KindAward:
		return "award"
	}
	return ""
}

// Make joins a kind and a base-36 id into a fullname (t3_abc123).
func Make(k Kind, id string) string { return string(k) + "_" + id }

// Split breaks a fullname into its kind and id. A value without a recognized
// "tN_" prefix returns ("", value), so a bare id passes through unchanged.
func Split(fullname string) (Kind, string) {
	fullname = strings.TrimSpace(fullname)
	if len(fullname) < 4 || fullname[0] != 't' || fullname[2] != '_' {
		return "", fullname
	}
	k := Kind(fullname[:2])
	if k.Name() == "" {
		return "", fullname
	}
	return k, fullname[3:]
}

// Classify inspects any Reddit URL, fullname, or bare id and returns a
// human-facing kind name ("post", "comment", "subreddit", "user") and the
// canonical id (a base-36 id for posts and comments, the name for subreddits and
// users). It returns ("", "") when the input is unrecognizable.
func Classify(s string) (name, id string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ""
	}

	// Fullname form: tN_id.
	if k, rest := Split(s); k != "" && !strings.Contains(s, "/") {
		return k.Name(), rest
	}

	low := strings.ToLower(s)
	switch {
	case strings.Contains(low, "redd.it/"):
		return "post", afterLast(s, "redd.it/")
	case strings.Contains(low, "/comments/"):
		// .../comments/<post>/<slug>/<comment>/ : a trailing segment past the
		// slug is a comment id, otherwise it is the post id.
		seg := segmentsAfter(s, "/comments/")
		if len(seg) >= 3 && seg[2] != "" {
			return "comment", seg[2]
		}
		if len(seg) >= 1 {
			return "post", seg[0]
		}
	case strings.Contains(low, "/r/"):
		return "subreddit", firstSegment(afterFold(s, "/r/"))
	case strings.Contains(low, "/user/"):
		return "user", firstSegment(afterFold(s, "/user/"))
	case strings.Contains(low, "/u/"):
		return "user", firstSegment(afterFold(s, "/u/"))
	}

	// Bare tokens: "r/golang", "u/spez", or a plain id treated as a post.
	switch {
	case strings.HasPrefix(low, "r/"):
		return "subreddit", firstSegment(s[2:])
	case strings.HasPrefix(low, "u/"):
		return "user", firstSegment(s[2:])
	case !strings.Contains(s, "/") && !strings.Contains(s, "."):
		return "post", s
	}
	return "", ""
}

// afterFold returns the part of s after the first case-insensitive match of sub.
func afterFold(s, sub string) string {
	i := strings.Index(strings.ToLower(s), strings.ToLower(sub))
	if i < 0 {
		return ""
	}
	return s[i+len(sub):]
}

// afterLast returns the part of s after the last match of sub, trimmed at the
// next path separator or query.
func afterLast(s, sub string) string {
	i := strings.LastIndex(s, sub)
	if i < 0 {
		return ""
	}
	return firstSegment(s[i+len(sub):])
}

// segmentsAfter returns the path segments following sub.
func segmentsAfter(s, sub string) []string {
	rest := afterFold(s, sub)
	rest = trimQuery(rest)
	parts := strings.Split(rest, "/")
	out := make([]string, 0, len(parts))
	out = append(out, parts...)
	return out
}

// firstSegment returns the first path segment of s (up to / ? or #).
func firstSegment(s string) string {
	s = trimQuery(s)
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}
	return s
}

func trimQuery(s string) string {
	if i := strings.IndexAny(s, "?#"); i >= 0 {
		return s[:i]
	}
	return s
}
