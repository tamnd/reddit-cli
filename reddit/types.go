// Package reddit reads public Reddit data: subreddit listings, posts, comment
// trees, user and subreddit profiles, search, rules, moderators, wiki pages, and
// the duplicate discussions of a link. It reads the public .json view that every
// Reddit path exposes and normalizes each "thing" into a rich record, with
// explicit empty/zero/[] for genuinely absent fields. It carries no CLI
// knowledge and depends only on the standard library, golang.org/x, and a
// pure-Go SQLite driver.
package reddit

import "time"

// Post is a Reddit link (a t3 thing): a self post or an outbound link.
type Post struct {
	PostID           string    `json:"post_id"`
	Fullname         string    `json:"fullname"`
	Subreddit        string    `json:"subreddit"`
	SubredditID      string    `json:"subreddit_id"`
	Title            string    `json:"title"`
	Author           string    `json:"author"`
	AuthorFullname   string    `json:"author_fullname"`
	Selftext         string    `json:"selftext"`
	URL              string    `json:"url"`
	Permalink        string    `json:"permalink"`
	Domain           string    `json:"domain"`
	IsSelf           bool      `json:"is_self"`
	IsVideo          bool      `json:"is_video"`
	Over18           bool      `json:"over_18"`
	Spoiler          bool      `json:"spoiler"`
	Stickied         bool      `json:"stickied"`
	Locked           bool      `json:"locked"`
	Score            int       `json:"score"`
	UpvoteRatio      float64   `json:"upvote_ratio"`
	Ups              int       `json:"ups"`
	NumComments      int       `json:"num_comments"`
	NumCrossposts    int       `json:"num_crossposts"`
	CreatedUTC       time.Time `json:"created_utc"`
	Edited           time.Time `json:"edited"`
	Gilded           int       `json:"gilded"`
	TotalAwards      int       `json:"total_awards"`
	LinkFlairText    string    `json:"link_flair_text"`
	AuthorFlairText  string    `json:"author_flair_text"`
	PostHint         string    `json:"post_hint"`
	Thumbnail        string    `json:"thumbnail"`
	Distinguished    string    `json:"distinguished"`
	RemovedCategory  string    `json:"removed_category"`
	GalleryImageURLs []string  `json:"gallery_image_urls"`
	CrosspostParent  string    `json:"crosspost_parent"`
	FetchedAt        time.Time `json:"fetched_at"`
}

// Comment is a Reddit comment (a t1 thing). One record per comment, with depth
// and parent links so a flattened tree keeps its shape.
type Comment struct {
	CommentID        string    `json:"comment_id"`
	Fullname         string    `json:"fullname"`
	LinkID           string    `json:"link_id"`
	ParentID         string    `json:"parent_id"`
	Subreddit        string    `json:"subreddit"`
	Author           string    `json:"author"`
	AuthorFullname   string    `json:"author_fullname"`
	Body             string    `json:"body"`
	Score            int       `json:"score"`
	Ups              int       `json:"ups"`
	Controversiality int       `json:"controversiality"`
	CreatedUTC       time.Time `json:"created_utc"`
	Edited           time.Time `json:"edited"`
	Depth            int       `json:"depth"`
	IsSubmitter      bool      `json:"is_submitter"`
	Stickied         bool      `json:"stickied"`
	Distinguished    string    `json:"distinguished"`
	Gilded           int       `json:"gilded"`
	TotalAwards      int       `json:"total_awards"`
	Permalink        string    `json:"permalink"`
	ReplyCount       int       `json:"reply_count"`
	MoreCount        int       `json:"more_count"`
	FetchedAt        time.Time `json:"fetched_at"`
}

// Subreddit is a community (a t5 thing).
type Subreddit struct {
	SubredditID        string    `json:"subreddit_id"`
	Name               string    `json:"name"`
	Title              string    `json:"title"`
	PublicDescription  string    `json:"public_description"`
	Description        string    `json:"description"`
	Subscribers        int64     `json:"subscribers"`
	ActiveUserCount    int64     `json:"active_user_count"`
	CreatedUTC         time.Time `json:"created_utc"`
	Over18             bool      `json:"over18"`
	Quarantine         bool      `json:"quarantine"`
	SubredditType      string    `json:"subreddit_type"`
	SubmissionType     string    `json:"submission_type"`
	Lang               string    `json:"lang"`
	URL                string    `json:"url"`
	CommunityIcon      string    `json:"community_icon"`
	IconImg            string    `json:"icon_img"`
	BannerImg          string    `json:"banner_img"`
	AdvertiserCategory string    `json:"advertiser_category"`
	FetchedAt          time.Time `json:"fetched_at"`
}

// User is a Reddit account (a t2 thing).
type User struct {
	UserID               string    `json:"user_id"`
	Name                 string    `json:"name"`
	CreatedUTC           time.Time `json:"created_utc"`
	LinkKarma            int64     `json:"link_karma"`
	CommentKarma         int64     `json:"comment_karma"`
	AwardeeKarma         int64     `json:"awardee_karma"`
	AwarderKarma         int64     `json:"awarder_karma"`
	TotalKarma           int64     `json:"total_karma"`
	IsGold               bool      `json:"is_gold"`
	IsMod                bool      `json:"is_mod"`
	IsEmployee           bool      `json:"is_employee"`
	Verified             bool      `json:"verified"`
	HasVerifiedEmail     bool      `json:"has_verified_email"`
	AcceptFollowers      bool      `json:"accept_followers"`
	IconImg              string    `json:"icon_img"`
	SubredditTitle       string    `json:"subreddit_title"`
	SubredditSubscribers int64     `json:"subreddit_subscribers"`
	URL                  string    `json:"url"`
	FetchedAt            time.Time `json:"fetched_at"`
}

// SubredditResult is a lean subreddit row from discovery or sr-typed search.
type SubredditResult struct {
	Position          int    `json:"position"`
	Name              string `json:"name"`
	Title             string `json:"title"`
	Subscribers       int64  `json:"subscribers"`
	PublicDescription string `json:"public_description"`
	Over18            bool   `json:"over18"`
	URL               string `json:"url"`
}

// UserResult is a lean user row from discovery or user-typed search.
type UserResult struct {
	Position   int       `json:"position"`
	Name       string    `json:"name"`
	TotalKarma int64     `json:"total_karma"`
	CreatedUTC time.Time `json:"created_utc"`
	URL        string    `json:"url"`
}

// Rule is one of a subreddit's posted rules.
type Rule struct {
	Subreddit       string    `json:"subreddit"`
	Kind            string    `json:"kind"`
	ShortName       string    `json:"short_name"`
	Description     string    `json:"description"`
	ViolationReason string    `json:"violation_reason"`
	Priority        int       `json:"priority"`
	CreatedUTC      time.Time `json:"created_utc"`
}

// Moderator is one of a subreddit's moderators.
type Moderator struct {
	Subreddit      string    `json:"subreddit"`
	Name           string    `json:"name"`
	AuthorFullname string    `json:"author_fullname"`
	ModPermissions []string  `json:"mod_permissions"`
	Date           time.Time `json:"date"`
}

// WikiPage is a subreddit wiki page with its markdown body and revision meta.
type WikiPage struct {
	Subreddit    string    `json:"subreddit"`
	Page         string    `json:"page"`
	ContentMD    string    `json:"content_md"`
	RevisionBy   string    `json:"revision_by"`
	RevisionDate time.Time `json:"revision_date"`
	MayRevise    bool      `json:"may_revise"`
	URL          string    `json:"url"`
}

// WikiIndexEntry is one page name from a subreddit's wiki page index.
type WikiIndexEntry struct {
	Subreddit string `json:"subreddit"`
	Page      string `json:"page"`
	URL       string `json:"url"`
}

// QueueItem is one row of the crawl queue.
type QueueItem struct {
	ID         int64  `json:"id"`
	URL        string `json:"url"`
	EntityType string `json:"entity_type"`
	Priority   int    `json:"priority"`
	HTMLPath   string `json:"html_path"`
}
