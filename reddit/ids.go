package reddit

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/tamnd/reddit-cli/pkg/thing"
)

// Classify is re-exported from pkg/thing so callers do not import both packages
// just to turn a URL or id into a (kind, id) pair.
func Classify(s string) (kind, id string) { return thing.Classify(s) }

// Sort is a subreddit listing sort.
type Sort string

const (
	SortHot           Sort = "hot"
	SortNew           Sort = "new"
	SortTop           Sort = "top"
	SortRising        Sort = "rising"
	SortControversial Sort = "controversial"
)

// ValidSort reports whether s is a listing sort reddit-cli accepts.
func ValidSort(s string) bool {
	switch Sort(s) {
	case SortHot, SortNew, SortTop, SortRising, SortControversial:
		return true
	}
	return false
}

// ValidTime reports whether t is a time window for top/controversial.
func ValidTime(t string) bool {
	switch t {
	case "hour", "day", "week", "month", "year", "all":
		return true
	}
	return false
}

// ListingParams are the shared pagination and windowing knobs.
type ListingParams struct {
	Sort  string
	Time  string // hour|day|week|month|year|all, for top/controversial
	Limit int    // per-page size, capped at MaxPageLimit
	After string // fullname cursor
	Count int    // running offset
}

func (p ListingParams) values() url.Values {
	v := url.Values{}
	limit := p.Limit
	if limit <= 0 || limit > MaxPageLimit {
		limit = MaxPageLimit
	}
	v.Set("limit", strconv.Itoa(limit))
	v.Set("raw_json", "1")
	if p.After != "" {
		v.Set("after", p.After)
	}
	if p.Count > 0 {
		v.Set("count", strconv.Itoa(p.Count))
	}
	if p.Time != "" {
		v.Set("t", p.Time)
	}
	return v
}

// SubredditListingURL builds a subreddit sort listing URL.
func SubredditListingURL(sub string, p ListingParams) string {
	sort := p.Sort
	if sort == "" {
		sort = string(SortHot)
	}
	return BaseURL + "/r/" + cleanName(sub) + "/" + sort + ".json?" + p.values().Encode()
}

// PostURL builds the post + comments JSON URL for a post id.
func PostURL(id string) string {
	v := url.Values{}
	v.Set("raw_json", "1")
	return BaseURL + "/comments/" + baseID(id) + ".json?" + v.Encode()
}

// CommentsURL builds the post + comments JSON URL with a sort and depth.
func CommentsURL(id, sort string, limit, depth int) string {
	v := url.Values{}
	v.Set("raw_json", "1")
	if sort != "" {
		v.Set("sort", sort)
	}
	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}
	if depth > 0 {
		v.Set("depth", strconv.Itoa(depth))
	}
	return BaseURL + "/comments/" + baseID(id) + ".json?" + v.Encode()
}

// SubredditAboutURL builds a subreddit about URL.
func SubredditAboutURL(sub string) string {
	return BaseURL + "/r/" + cleanName(sub) + "/about.json?raw_json=1"
}

// UserAboutURL builds a user about URL.
func UserAboutURL(name string) string {
	return BaseURL + "/user/" + cleanName(name) + "/about.json?raw_json=1"
}

// UserListingURL builds a user submitted/comments/overview listing URL.
func UserListingURL(name, kind string, p ListingParams) string {
	if kind == "" {
		kind = "overview"
	}
	return BaseURL + "/user/" + cleanName(name) + "/" + kind + ".json?" + p.values().Encode()
}

// SearchURL builds a search URL. When sub is non-empty the search is restricted
// to that subreddit.
func SearchURL(query, sub string, p ListingParams, typ string, nsfw bool) string {
	v := p.values()
	v.Set("q", query)
	if p.Sort != "" {
		v.Set("sort", p.Sort)
	}
	if typ != "" {
		v.Set("type", typ)
	}
	if nsfw {
		v.Set("include_over_18", "1")
	}
	if sub != "" {
		v.Set("restrict_sr", "1")
		return BaseURL + "/r/" + cleanName(sub) + "/search.json?" + v.Encode()
	}
	return BaseURL + "/search.json?" + v.Encode()
}

// SubredditsSearchURL builds the subreddit discovery search URL.
func SubredditsSearchURL(query string, p ListingParams) string {
	v := p.values()
	v.Set("q", query)
	return BaseURL + "/subreddits/search.json?" + v.Encode()
}

// UsersSearchURL builds the user discovery search URL.
func UsersSearchURL(query string, p ListingParams) string {
	v := p.values()
	v.Set("q", query)
	return BaseURL + "/users/search.json?" + v.Encode()
}

// RulesURL builds a subreddit rules URL.
func RulesURL(sub string) string {
	return BaseURL + "/r/" + cleanName(sub) + "/about/rules.json?raw_json=1"
}

// ModeratorsURL builds a subreddit moderators URL.
func ModeratorsURL(sub string) string {
	return BaseURL + "/r/" + cleanName(sub) + "/about/moderators.json?raw_json=1"
}

// WikiURL builds a subreddit wiki page URL.
func WikiURL(sub, page string) string {
	if page == "" {
		page = "index"
	}
	return BaseURL + "/r/" + cleanName(sub) + "/wiki/" + page + ".json?raw_json=1"
}

// WikiPagesURL builds the wiki page index URL.
func WikiPagesURL(sub string) string {
	return BaseURL + "/r/" + cleanName(sub) + "/wiki/pages.json?raw_json=1"
}

// DuplicatesURL builds the duplicates URL for a post id.
func DuplicatesURL(id string) string {
	return BaseURL + "/duplicates/" + baseID(id) + ".json?raw_json=1"
}

// MoreChildrenURL builds the comment-tree expansion URL for a batch of child
// ids under a link.
func MoreChildrenURL(linkFullname, sort string, children []string) string {
	v := url.Values{}
	v.Set("api_type", "json")
	v.Set("raw_json", "1")
	v.Set("link_id", linkFullname)
	v.Set("children", strings.Join(children, ","))
	if sort != "" {
		v.Set("sort", sort)
	}
	return BaseURL + "/api/morechildren.json?" + v.Encode()
}

// PermalinkURL turns a relative reddit permalink into an absolute URL.
func PermalinkURL(permalink string) string {
	if permalink == "" {
		return ""
	}
	if strings.HasPrefix(permalink, "http") {
		return permalink
	}
	return BaseURL + permalink
}

// ResolveURL turns any accepted id or URL into the page a human would open.
func ResolveURL(arg string) string {
	if strings.HasPrefix(arg, "http") {
		return arg
	}
	kind, id := thing.Classify(arg)
	switch kind {
	case "post":
		return BaseURL + "/comments/" + id + "/"
	case "subreddit":
		return BaseURL + "/r/" + id + "/"
	case "user":
		return BaseURL + "/user/" + id + "/"
	}
	return ""
}

// baseID strips a tN_ prefix and a trailing slug so a bare base-36 id remains.
func baseID(s string) string {
	_, id := thing.Classify(s)
	if id != "" {
		return id
	}
	return strings.TrimSpace(s)
}

// cleanName strips r/, u/, /r/, /user/ wrappers and returns the bare name.
func cleanName(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "http") || strings.Contains(s, "/") {
		if k, id := thing.Classify(s); (k == "subreddit" || k == "user") && id != "" {
			return id
		}
	}
	s = strings.TrimPrefix(s, "/")
	for _, p := range []string{"r/", "u/", "user/"} {
		if strings.HasPrefix(strings.ToLower(s), p) {
			return strings.SplitN(s[len(p):], "/", 2)[0]
		}
	}
	return s
}
