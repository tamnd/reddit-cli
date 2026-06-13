package reddit

import (
	"encoding/json"
	"testing"
)

func TestParsePost(t *testing.T) {
	const data = `{
		"id": "1abc23",
		"name": "t3_1abc23",
		"subreddit": "golang",
		"subreddit_id": "t5_2rc7j",
		"title": "Generics in Go",
		"author": "gopher",
		"author_fullname": "t2_xyz",
		"selftext": "body text",
		"url": "https://example.com/x",
		"permalink": "/r/golang/comments/1abc23/generics_in_go/",
		"domain": "example.com",
		"is_self": false,
		"over_18": false,
		"is_original_content": true,
		"pinned": false,
		"archived": true,
		"subreddit_subscribers": 285000,
		"score": 142,
		"upvote_ratio": 0.97,
		"num_comments": 23,
		"created_utc": 1700000000,
		"edited": false,
		"link_flair_text": "news",
		"media": {"reddit_video": {"fallback_url": "https://v.redd.it/abc/DASH_720.mp4"}},
		"preview": {"images": [{"source": {"url": "https://preview.redd.it/abc.jpg?width=640"}}]}
	}`
	p, err := parsePost(json.RawMessage(data))
	if err != nil {
		t.Fatalf("parsePost: %v", err)
	}
	if p.PostID != "1abc23" || p.Fullname != "t3_1abc23" {
		t.Errorf("id/fullname = %q/%q", p.PostID, p.Fullname)
	}
	if p.Subreddit != "golang" || p.Score != 142 || p.NumComments != 23 {
		t.Errorf("subreddit/score/comments = %q/%d/%d", p.Subreddit, p.Score, p.NumComments)
	}
	if p.Permalink != BaseURL+"/r/golang/comments/1abc23/generics_in_go/" {
		t.Errorf("permalink not absolutized: %q", p.Permalink)
	}
	if !p.Edited.IsZero() {
		t.Errorf("edited should be zero for false, got %v", p.Edited)
	}
	if p.CreatedUTC.IsZero() {
		t.Error("created_utc should be set")
	}
	if !p.IsOriginalContent || !p.Archived || p.Pinned {
		t.Errorf("oc/archived/pinned = %v/%v/%v", p.IsOriginalContent, p.Archived, p.Pinned)
	}
	if p.SubredditSubscribers != 285000 {
		t.Errorf("subreddit_subscribers = %d", p.SubredditSubscribers)
	}
	if p.MediaURL != "https://v.redd.it/abc/DASH_720.mp4" {
		t.Errorf("media_url = %q", p.MediaURL)
	}
	if p.PreviewImageURL != "https://preview.redd.it/abc.jpg?width=640" {
		t.Errorf("preview_image_url = %q", p.PreviewImageURL)
	}
}

func TestParsePostEditedTimestamp(t *testing.T) {
	p, err := parsePost(json.RawMessage(`{"id":"x","edited":1700000500,"created_utc":1700000000}`))
	if err != nil {
		t.Fatal(err)
	}
	if p.Edited.IsZero() {
		t.Error("edited should be set when a timestamp is present")
	}
}

func TestParseComment(t *testing.T) {
	const data = `{
		"id": "c1",
		"name": "t1_c1",
		"link_id": "t3_1abc23",
		"parent_id": "t3_1abc23",
		"subreddit": "golang",
		"author": "gopher",
		"body": "nice post",
		"score": 9,
		"created_utc": 1700000100,
		"edited": false,
		"depth": 0,
		"collapsed": true,
		"score_hidden": false,
		"author_flair_text": "Gopher",
		"replies": ""
	}`
	cm, replies, err := parseComment(json.RawMessage(data))
	if err != nil {
		t.Fatalf("parseComment: %v", err)
	}
	if cm.CommentID != "c1" || cm.Fullname != "t1_c1" {
		t.Errorf("id/fullname = %q/%q", cm.CommentID, cm.Fullname)
	}
	if cm.Body != "nice post" || cm.Score != 9 {
		t.Errorf("body/score = %q/%d", cm.Body, cm.Score)
	}
	if !cm.Collapsed || cm.ScoreHidden || cm.AuthorFlairText != "Gopher" {
		t.Errorf("collapsed/hidden/flair = %v/%v/%q", cm.Collapsed, cm.ScoreHidden, cm.AuthorFlairText)
	}
	// Reddit sends "" (an empty JSON string) for a comment with no replies;
	// summarizing it must not panic and must report no replies.
	if rc, mc := summarizeReplies(replies); rc != 0 || mc != 0 {
		t.Errorf("summarizeReplies on empty = %d/%d, want 0/0", rc, mc)
	}
}

func TestParseSubreddit(t *testing.T) {
	const data = `{
		"name": "t5_2rc7j",
		"display_name": "golang",
		"title": "The Go Programming Language",
		"public_description": "Ask questions and post articles about Go.",
		"subscribers": 285000,
		"created_utc": 1234567890,
		"over18": false,
		"wiki_enabled": true,
		"subreddit_type": "public",
		"url": "/r/golang/",
		"icon_img": "https://b.thumbs.redditmedia.com/icon.png?width=256"
	}`
	s, err := parseSubreddit(json.RawMessage(data))
	if err != nil {
		t.Fatalf("parseSubreddit: %v", err)
	}
	if s.Name != "golang" || s.SubredditID != "t5_2rc7j" {
		t.Errorf("name/id = %q/%q", s.Name, s.SubredditID)
	}
	if s.Subscribers != 285000 {
		t.Errorf("subscribers = %d", s.Subscribers)
	}
	if !s.WikiEnabled {
		t.Errorf("wiki_enabled = %v", s.WikiEnabled)
	}
	if s.URL != BaseURL+"/r/golang/" {
		t.Errorf("url not absolutized: %q", s.URL)
	}
	if s.IconImg != "https://b.thumbs.redditmedia.com/icon.png" {
		t.Errorf("icon query not stripped: %q", s.IconImg)
	}
}

func TestParseUser(t *testing.T) {
	const data = `{
		"id": "xyz",
		"name": "gopher",
		"created_utc": 1300000000,
		"link_karma": 1200,
		"comment_karma": 3400,
		"total_karma": 4600,
		"is_gold": true,
		"verified": true,
		"subreddit": {"title": "gopher", "public_description": "Go all day.", "subscribers": 42}
	}`
	u, err := parseUser(json.RawMessage(data))
	if err != nil {
		t.Fatalf("parseUser: %v", err)
	}
	if u.Name != "gopher" || u.UserID != "t2_xyz" {
		t.Errorf("name/id = %q/%q", u.Name, u.UserID)
	}
	if u.TotalKarma != 4600 || !u.IsGold {
		t.Errorf("karma/gold = %d/%v", u.TotalKarma, u.IsGold)
	}
	if u.SubredditDescription != "Go all day." || u.SubredditSubscribers != 42 {
		t.Errorf("profile desc/subs = %q/%d", u.SubredditDescription, u.SubredditSubscribers)
	}
	if u.URL != BaseURL+"/user/gopher/" {
		t.Errorf("url = %q", u.URL)
	}
}

func TestParsePostsFiltersKinds(t *testing.T) {
	children := []thingEnvelope{
		{Kind: "t3", Data: json.RawMessage(`{"id":"a"}`)},
		{Kind: "t1", Data: json.RawMessage(`{"id":"b"}`)},
		{Kind: "t3", Data: json.RawMessage(`{"id":"c"}`)},
	}
	posts := parsePosts(children)
	if len(posts) != 2 {
		t.Fatalf("expected 2 t3 posts, got %d", len(posts))
	}
	if posts[0].PostID != "a" || posts[1].PostID != "c" {
		t.Errorf("ids = %q,%q", posts[0].PostID, posts[1].PostID)
	}
}
