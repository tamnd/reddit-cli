package reddit

import (
	"encoding/json"
	"strings"
	"time"
)

// epoch turns a Unix-seconds float into a UTC time.Time, zero for 0.
func epoch(f float64) time.Time {
	if f <= 0 {
		return time.Time{}
	}
	sec := int64(f)
	nsec := int64((f - float64(sec)) * 1e9)
	return time.Unix(sec, nsec).UTC()
}

// editedTime decodes Reddit's edited union, which is false when never edited and
// a Unix-seconds float otherwise.
func editedTime(raw json.RawMessage) time.Time {
	if len(raw) == 0 {
		return time.Time{}
	}
	var f float64
	if json.Unmarshal(raw, &f) == nil {
		return epoch(f)
	}
	return time.Time{}
}

// postData mirrors the fields reddit-cli reads off a t3 link.
type postData struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	Subreddit            string          `json:"subreddit"`
	SubredditID          string          `json:"subreddit_id"`
	Title                string          `json:"title"`
	Author               string          `json:"author"`
	AuthorFullname       string          `json:"author_fullname"`
	Selftext             string          `json:"selftext"`
	URL                  string          `json:"url"`
	Permalink            string          `json:"permalink"`
	Domain               string          `json:"domain"`
	IsSelf               bool            `json:"is_self"`
	IsVideo              bool            `json:"is_video"`
	IsOriginalContent    bool            `json:"is_original_content"`
	Over18               bool            `json:"over_18"`
	Spoiler              bool            `json:"spoiler"`
	Stickied             bool            `json:"stickied"`
	Pinned               bool            `json:"pinned"`
	Locked               bool            `json:"locked"`
	Archived             bool            `json:"archived"`
	Score                int             `json:"score"`
	UpvoteRatio          float64         `json:"upvote_ratio"`
	Ups                  int             `json:"ups"`
	NumComments          int             `json:"num_comments"`
	NumCrossposts        int             `json:"num_crossposts"`
	SubredditSubscribers int64           `json:"subreddit_subscribers"`
	CreatedUTC           float64         `json:"created_utc"`
	Edited               json.RawMessage `json:"edited"`
	Gilded               int             `json:"gilded"`
	TotalAwards          int             `json:"total_awards_received"`
	LinkFlairText        string          `json:"link_flair_text"`
	AuthorFlairText      string          `json:"author_flair_text"`
	PostHint             string          `json:"post_hint"`
	Thumbnail            string          `json:"thumbnail"`
	Distinguished        string          `json:"distinguished"`
	RemovedCategory      string          `json:"removed_by_category"`
	CrosspostParent      string          `json:"crosspost_parent"`
	Media                struct {
		RedditVideo struct {
			FallbackURL string `json:"fallback_url"`
		} `json:"reddit_video"`
	} `json:"media"`
	Preview struct {
		Images []struct {
			Source struct {
				URL string `json:"url"`
			} `json:"source"`
		} `json:"images"`
	} `json:"preview"`
	MediaMetadata map[string]struct {
		S struct {
			U string `json:"u"`
		} `json:"s"`
	} `json:"media_metadata"`
	GalleryData struct {
		Items []struct {
			MediaID string `json:"media_id"`
		} `json:"items"`
	} `json:"gallery_data"`
}

// parsePost decodes a t3 data payload into a Post.
func parsePost(data json.RawMessage) (Post, error) {
	var d postData
	if err := json.Unmarshal(data, &d); err != nil {
		return Post{}, err
	}
	p := Post{
		PostID:               d.ID,
		Fullname:             orFullname(d.Name, "t3", d.ID),
		Subreddit:            d.Subreddit,
		SubredditID:          d.SubredditID,
		Title:                d.Title,
		Author:               d.Author,
		AuthorFullname:       d.AuthorFullname,
		Selftext:             d.Selftext,
		URL:                  d.URL,
		Permalink:            PermalinkURL(d.Permalink),
		Domain:               d.Domain,
		IsSelf:               d.IsSelf,
		IsVideo:              d.IsVideo,
		IsOriginalContent:    d.IsOriginalContent,
		Over18:               d.Over18,
		Spoiler:              d.Spoiler,
		Stickied:             d.Stickied,
		Pinned:               d.Pinned,
		Locked:               d.Locked,
		Archived:             d.Archived,
		Score:                d.Score,
		UpvoteRatio:          d.UpvoteRatio,
		Ups:                  d.Ups,
		NumComments:          d.NumComments,
		NumCrossposts:        d.NumCrossposts,
		SubredditSubscribers: d.SubredditSubscribers,
		CreatedUTC:           epoch(d.CreatedUTC),
		Edited:               editedTime(d.Edited),
		Gilded:               d.Gilded,
		TotalAwards:          d.TotalAwards,
		LinkFlairText:        d.LinkFlairText,
		AuthorFlairText:      d.AuthorFlairText,
		PostHint:             d.PostHint,
		Thumbnail:            d.Thumbnail,
		MediaURL:             d.Media.RedditVideo.FallbackURL,
		PreviewImageURL:      previewImageURL(d),
		Distinguished:        d.Distinguished,
		RemovedCategory:      d.RemovedCategory,
		CrosspostParent:      d.CrosspostParent,
		GalleryImageURLs:     galleryURLs(d),
		FetchedAt:            time.Now().UTC(),
	}
	return p, nil
}

// previewImageURL returns the source URL of a post's first preview image, with
// its HTML-escaped query left intact so the link resolves as Reddit serves it.
func previewImageURL(d postData) string {
	if len(d.Preview.Images) == 0 {
		return ""
	}
	return d.Preview.Images[0].Source.URL
}

// galleryURLs resolves a gallery post's ordered image URLs from its metadata.
func galleryURLs(d postData) []string {
	if len(d.GalleryData.Items) == 0 || len(d.MediaMetadata) == 0 {
		return nil
	}
	var out []string
	for _, it := range d.GalleryData.Items {
		if m, ok := d.MediaMetadata[it.MediaID]; ok && m.S.U != "" {
			out = append(out, m.S.U)
		}
	}
	return out
}

// commentData mirrors the fields reddit-cli reads off a t1 comment.
type commentData struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	LinkID           string          `json:"link_id"`
	ParentID         string          `json:"parent_id"`
	Subreddit        string          `json:"subreddit"`
	Author           string          `json:"author"`
	AuthorFullname   string          `json:"author_fullname"`
	Body             string          `json:"body"`
	Score            int             `json:"score"`
	Ups              int             `json:"ups"`
	Controversiality int             `json:"controversiality"`
	CreatedUTC       float64         `json:"created_utc"`
	Edited           json.RawMessage `json:"edited"`
	Depth            int             `json:"depth"`
	IsSubmitter      bool            `json:"is_submitter"`
	Stickied         bool            `json:"stickied"`
	ScoreHidden      bool            `json:"score_hidden"`
	Collapsed        bool            `json:"collapsed"`
	Distinguished    string          `json:"distinguished"`
	AuthorFlairText  string          `json:"author_flair_text"`
	Gilded           int             `json:"gilded"`
	TotalAwards      int             `json:"total_awards_received"`
	Permalink        string          `json:"permalink"`
	Replies          json.RawMessage `json:"replies"`
}

// subredditData mirrors the fields reddit-cli reads off a t5 subreddit.
type subredditData struct {
	Name               string  `json:"name"`
	DisplayName        string  `json:"display_name"`
	Title              string  `json:"title"`
	PublicDescription  string  `json:"public_description"`
	Description        string  `json:"description"`
	Subscribers        int64   `json:"subscribers"`
	ActiveUserCount    int64   `json:"active_user_count"`
	CreatedUTC         float64 `json:"created_utc"`
	Over18             bool    `json:"over18"`
	Quarantine         bool    `json:"quarantine"`
	WikiEnabled        bool    `json:"wiki_enabled"`
	SubredditType      string  `json:"subreddit_type"`
	SubmissionType     string  `json:"submission_type"`
	Lang               string  `json:"lang"`
	URL                string  `json:"url"`
	CommunityIcon      string  `json:"community_icon"`
	IconImg            string  `json:"icon_img"`
	BannerImg          string  `json:"banner_img"`
	AdvertiserCategory string  `json:"advertiser_category"`
}

func parseSubreddit(data json.RawMessage) (Subreddit, error) {
	var d subredditData
	if err := json.Unmarshal(data, &d); err != nil {
		return Subreddit{}, err
	}
	return Subreddit{
		SubredditID:        d.Name,
		Name:               d.DisplayName,
		Title:              d.Title,
		PublicDescription:  d.PublicDescription,
		Description:        d.Description,
		Subscribers:        d.Subscribers,
		ActiveUserCount:    d.ActiveUserCount,
		CreatedUTC:         epoch(d.CreatedUTC),
		Over18:             d.Over18,
		Quarantine:         d.Quarantine,
		WikiEnabled:        d.WikiEnabled,
		SubredditType:      d.SubredditType,
		SubmissionType:     d.SubmissionType,
		Lang:               d.Lang,
		URL:                PermalinkURL(d.URL),
		CommunityIcon:      stripQuery(d.CommunityIcon),
		IconImg:            stripQuery(d.IconImg),
		BannerImg:          stripQuery(d.BannerImg),
		AdvertiserCategory: d.AdvertiserCategory,
		FetchedAt:          time.Now().UTC(),
	}, nil
}

// userData mirrors the fields reddit-cli reads off a t2 account.
type userData struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	CreatedUTC       float64 `json:"created_utc"`
	LinkKarma        int64   `json:"link_karma"`
	CommentKarma     int64   `json:"comment_karma"`
	AwardeeKarma     int64   `json:"awardee_karma"`
	AwarderKarma     int64   `json:"awarder_karma"`
	TotalKarma       int64   `json:"total_karma"`
	IsGold           bool    `json:"is_gold"`
	IsMod            bool    `json:"is_mod"`
	IsEmployee       bool    `json:"is_employee"`
	Verified         bool    `json:"verified"`
	HasVerifiedEmail bool    `json:"has_verified_email"`
	AcceptFollowers  bool    `json:"accept_followers"`
	IconImg          string  `json:"icon_img"`
	Subreddit        struct {
		Title             string `json:"title"`
		PublicDescription string `json:"public_description"`
		Subscribers       int64  `json:"subscribers"`
	} `json:"subreddit"`
}

func parseUser(data json.RawMessage) (User, error) {
	var d userData
	if err := json.Unmarshal(data, &d); err != nil {
		return User{}, err
	}
	return User{
		UserID:               orFullname("", "t2", d.ID),
		Name:                 d.Name,
		CreatedUTC:           epoch(d.CreatedUTC),
		LinkKarma:            d.LinkKarma,
		CommentKarma:         d.CommentKarma,
		AwardeeKarma:         d.AwardeeKarma,
		AwarderKarma:         d.AwarderKarma,
		TotalKarma:           d.TotalKarma,
		IsGold:               d.IsGold,
		IsMod:                d.IsMod,
		IsEmployee:           d.IsEmployee,
		Verified:             d.Verified,
		HasVerifiedEmail:     d.HasVerifiedEmail,
		AcceptFollowers:      d.AcceptFollowers,
		IconImg:              stripQuery(d.IconImg),
		SubredditTitle:       d.Subreddit.Title,
		SubredditDescription: d.Subreddit.PublicDescription,
		SubredditSubscribers: d.Subreddit.Subscribers,
		URL:                  BaseURL + "/user/" + d.Name + "/",
		FetchedAt:            time.Now().UTC(),
	}, nil
}

// parseChildren turns a listing's children into Post records, skipping
// non-t3 things (a mixed overview keeps only links here).
func parsePosts(children []thingEnvelope) []Post {
	var out []Post
	for _, ch := range children {
		if ch.Kind != string(kindLink) {
			continue
		}
		if p, err := parsePost(ch.Data); err == nil {
			out = append(out, p)
		}
	}
	return out
}

// parseCommentsFlat turns listing children into flat Comment records, skipping
// non-t1 things (overview keeps only comments here).
func parseCommentsFlat(children []thingEnvelope) []Comment {
	var out []Comment
	for _, ch := range children {
		if ch.Kind != string(kindComment) {
			continue
		}
		if cm, _, err := parseComment(ch.Data); err == nil {
			out = append(out, cm)
		}
	}
	return out
}

// parseComment decodes a t1 data payload into a Comment, returning the raw
// replies listing for the tree walker to recurse into.
func parseComment(data json.RawMessage) (Comment, json.RawMessage, error) {
	var d commentData
	if err := json.Unmarshal(data, &d); err != nil {
		return Comment{}, nil, err
	}
	cm := Comment{
		CommentID:        d.ID,
		Fullname:         orFullname(d.Name, "t1", d.ID),
		LinkID:           d.LinkID,
		ParentID:         d.ParentID,
		Subreddit:        d.Subreddit,
		Author:           d.Author,
		AuthorFullname:   d.AuthorFullname,
		Body:             d.Body,
		Score:            d.Score,
		Ups:              d.Ups,
		Controversiality: d.Controversiality,
		CreatedUTC:       epoch(d.CreatedUTC),
		Edited:           editedTime(d.Edited),
		Depth:            d.Depth,
		IsSubmitter:      d.IsSubmitter,
		Stickied:         d.Stickied,
		ScoreHidden:      d.ScoreHidden,
		Collapsed:        d.Collapsed,
		Distinguished:    d.Distinguished,
		AuthorFlairText:  d.AuthorFlairText,
		Gilded:           d.Gilded,
		TotalAwards:      d.TotalAwards,
		Permalink:        PermalinkURL(d.Permalink),
		FetchedAt:        time.Now().UTC(),
	}
	return cm, d.Replies, nil
}

const (
	kindComment   = "t1"
	kindLink      = "t3"
	kindSubreddit = "t5"
)

// orFullname returns name when set, otherwise builds kind_id (when id is set).
func orFullname(name, kind, id string) string {
	if name != "" {
		return name
	}
	if id == "" {
		return ""
	}
	return kind + "_" + id
}

func stripQuery(s string) string {
	if i := strings.IndexByte(s, '?'); i >= 0 {
		return s[:i]
	}
	return s
}
