package reddit

import (
	"net/url"
	"strings"
	"testing"
)

func TestSubredditListingURL(t *testing.T) {
	got := SubredditListingURL("r/golang", ListingParams{Sort: "top", Time: "week", Limit: 25})
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if u.Path != "/r/golang/top.json" {
		t.Errorf("path = %q", u.Path)
	}
	q := u.Query()
	if q.Get("limit") != "25" || q.Get("t") != "week" || q.Get("raw_json") != "1" {
		t.Errorf("query = %v", q)
	}
}

func TestSubredditListingURLDefaultSort(t *testing.T) {
	got := SubredditListingURL("golang", ListingParams{})
	if !strings.Contains(got, "/r/golang/hot.json") {
		t.Errorf("default sort should be hot: %q", got)
	}
}

func TestListingParamsCapsLimit(t *testing.T) {
	v := ListingParams{Limit: 9999}.values()
	if v.Get("limit") != "100" {
		t.Errorf("limit should cap at 100, got %q", v.Get("limit"))
	}
}

func TestPostAndCommentsURL(t *testing.T) {
	if got := PostURL("t3_1abc23"); !strings.Contains(got, "/comments/1abc23.json") {
		t.Errorf("PostURL = %q", got)
	}
	got := CommentsURL("https://www.reddit.com/r/golang/comments/1abc23/title/", "top", 50, 3)
	u, _ := url.Parse(got)
	if u.Path != "/comments/1abc23.json" {
		t.Errorf("path = %q", u.Path)
	}
	q := u.Query()
	if q.Get("sort") != "top" || q.Get("limit") != "50" || q.Get("depth") != "3" {
		t.Errorf("query = %v", q)
	}
}

func TestSearchURLRestrict(t *testing.T) {
	got := SearchURL("generics", "golang", ListingParams{Limit: 10}, "link", false)
	u, _ := url.Parse(got)
	if u.Path != "/r/golang/search.json" {
		t.Errorf("path = %q", u.Path)
	}
	if u.Query().Get("restrict_sr") != "1" || u.Query().Get("q") != "generics" {
		t.Errorf("query = %v", u.Query())
	}
}

func TestMoreChildrenURL(t *testing.T) {
	got := MoreChildrenURL("t3_1abc23", "top", []string{"x1", "x2"})
	u, _ := url.Parse(got)
	q := u.Query()
	if q.Get("link_id") != "t3_1abc23" || q.Get("children") != "x1,x2" || q.Get("api_type") != "json" {
		t.Errorf("query = %v", q)
	}
}

func TestResolveURL(t *testing.T) {
	cases := map[string]string{
		"t5_2rc7j":    BaseURL + "/r/2rc7j/",
		"u/spez":      BaseURL + "/user/spez/",
		"t3_1abc23":   BaseURL + "/comments/1abc23/",
		"https://x/y": "https://x/y",
	}
	for in, want := range cases {
		if got := ResolveURL(in); got != want {
			t.Errorf("ResolveURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCleanName(t *testing.T) {
	cases := map[string]string{
		"golang":                            "golang",
		"r/golang":                          "golang",
		"/r/golang":                         "golang",
		"u/spez":                            "spez",
		"https://www.reddit.com/r/golang/":  "golang",
		"https://www.reddit.com/user/spez/": "spez",
	}
	for in, want := range cases {
		if got := cleanName(in); got != want {
			t.Errorf("cleanName(%q) = %q, want %q", in, got, want)
		}
	}
}
