package thing

import "testing"

func TestSplitAndMake(t *testing.T) {
	cases := []struct {
		full string
		kind Kind
		id   string
	}{
		{"t3_abc123", KindLink, "abc123"},
		{"t1_def456", KindComment, "def456"},
		{"t5_golang", KindSubreddit, "golang"},
		{"t2_xyz", KindAccount, "xyz"},
		{"abc123", "", "abc123"},   // bare id passes through
		{"t9_nope", "", "t9_nope"}, // unknown kind passes through
	}
	for _, c := range cases {
		k, id := Split(c.full)
		if k != c.kind || id != c.id {
			t.Errorf("Split(%q) = (%q,%q), want (%q,%q)", c.full, k, id, c.kind, c.id)
		}
		if c.kind != "" {
			if got := Make(c.kind, c.id); got != c.full {
				t.Errorf("Make(%q,%q) = %q, want %q", c.kind, c.id, got, c.full)
			}
		}
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		in   string
		name string
		id   string
	}{
		{"abc123", "post", "abc123"},
		{"t3_abc123", "post", "abc123"},
		{"t1_def456", "comment", "def456"},
		{"r/golang", "subreddit", "golang"},
		{"/r/golang", "subreddit", "golang"},
		{"u/spez", "user", "spez"},
		{"golang", "post", "golang"}, // a bare word is treated as a post id
		{"https://www.reddit.com/r/golang/comments/abc123/some_title/", "post", "abc123"},
		{"https://www.reddit.com/r/golang/comments/abc123/some_title/def456/", "comment", "def456"},
		{"https://old.reddit.com/r/Programming/", "subreddit", "Programming"},
		{"https://www.reddit.com/user/spez/", "user", "spez"},
		{"https://redd.it/abc123", "post", "abc123"},
		{"", "", ""},
	}
	for _, c := range cases {
		name, id := Classify(c.in)
		if name != c.name || id != c.id {
			t.Errorf("Classify(%q) = (%q,%q), want (%q,%q)", c.in, name, id, c.name, c.id)
		}
	}
}
