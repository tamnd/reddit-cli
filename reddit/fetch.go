package reddit

import (
	"context"
	"encoding/json"
	"fmt"
)

// fetchJSON gets a URL and returns its body, mapping a non-200 to the right
// sentinel error (not found, blocked, private, banned).
func (c *Client) fetchJSON(ctx context.Context, url string) ([]byte, error) {
	body, code, err := c.cachedFetch(ctx, url)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		if e := classifyErrorBody(code, body); e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("unexpected HTTP %d", code)
	}
	return body, nil
}

// Subreddit fetches a subreddit's about record.
func (c *Client) Subreddit(ctx context.Context, name string) (*Subreddit, error) {
	body, err := c.fetchJSON(ctx, SubredditAboutURL(name))
	if err != nil {
		return nil, err
	}
	var env thingEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, err
	}
	if env.Kind != string(kindSubreddit) {
		return nil, fmt.Errorf("expected a subreddit, got %q", env.Kind)
	}
	s, err := parseSubreddit(env.Data)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// User fetches a user's account record.
func (c *Client) User(ctx context.Context, name string) (*User, error) {
	body, err := c.fetchJSON(ctx, UserAboutURL(name))
	if err != nil {
		return nil, err
	}
	var env thingEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, err
	}
	if env.Kind != "t2" {
		return nil, fmt.Errorf("expected a user, got %q", env.Kind)
	}
	u, err := parseUser(env.Data)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Posts walks a subreddit's sort listing into Post records.
func (c *Client) Posts(ctx context.Context, sub string, p ListingParams, pages int) ([]Post, error) {
	var out []Post
	err := c.walkListing(ctx,
		func(after string, count int) string {
			pp := p
			pp.After = after
			pp.Count = count
			return SubredditListingURL(sub, pp)
		}, pages, p.Limit,
		func(children []thingEnvelope) (int, error) {
			out = append(out, parsePosts(children)...)
			return len(out), nil
		})
	if err != nil {
		return out, err
	}
	return trimPosts(out, p.Limit), nil
}

// Post fetches a single post (the link), ignoring its comments.
func (c *Client) Post(ctx context.Context, id string) (*Post, error) {
	body, err := c.fetchJSON(ctx, PostURL(id))
	if err != nil {
		return nil, err
	}
	post, _, err := decodePostAndComments(body)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrNotFound
	}
	return post, nil
}

// Comments fetches a post's comment tree, flattened into Comment records. When
// expandMore is set, the collapsed "load more" stubs are expanded via the
// morechildren endpoint, bounded by an internal round cap.
func (c *Client) Comments(ctx context.Context, id, sort string, limit, depth int, expand bool) ([]Comment, error) {
	body, err := c.fetchJSON(ctx, CommentsURL(id, sort, limit, depth))
	if err != nil {
		return nil, err
	}
	post, commentChildren, err := decodePostAndComments(body)
	if err != nil {
		return nil, err
	}
	var out []Comment
	var pending []moreData
	flattenTree(commentChildren, &out, &pending)
	if expand && post != nil && len(pending) > 0 {
		const rounds = 8
		if err := c.expandMore(ctx, post.Fullname, sort, pending, &out, rounds); err != nil {
			return out, err
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// decodePostAndComments splits the two-element array /comments/<id>.json
// returns: the post (from listing 0) and the comment children (from listing 1).
func decodePostAndComments(body []byte) (*Post, []thingEnvelope, error) {
	var arr []json.RawMessage
	if err := json.Unmarshal(body, &arr); err != nil {
		return nil, nil, fmt.Errorf("decode comments array: %w", err)
	}
	if len(arr) < 1 {
		return nil, nil, ErrNotFound
	}
	postLd, err := decodeListing(arr[0])
	if err != nil {
		return nil, nil, err
	}
	var post *Post
	if len(postLd.Children) > 0 && postLd.Children[0].Kind == string(kindLink) {
		if p, err := parsePost(postLd.Children[0].Data); err == nil {
			post = &p
		}
	}
	var commentChildren []thingEnvelope
	if len(arr) >= 2 {
		if cld, err := decodeListing(arr[1]); err == nil {
			commentChildren = cld.Children
		}
	}
	return post, commentChildren, nil
}

// UserPosts walks a user's submitted posts.
func (c *Client) UserPosts(ctx context.Context, name string, p ListingParams, pages int) ([]Post, error) {
	var out []Post
	err := c.walkListing(ctx,
		func(after string, count int) string {
			pp := p
			pp.After = after
			pp.Count = count
			return UserListingURL(name, "submitted", pp)
		}, pages, p.Limit,
		func(children []thingEnvelope) (int, error) {
			out = append(out, parsePosts(children)...)
			return len(out), nil
		})
	if err != nil {
		return out, err
	}
	return trimPosts(out, p.Limit), nil
}

// UserComments walks a user's comments.
func (c *Client) UserComments(ctx context.Context, name string, p ListingParams, pages int) ([]Comment, error) {
	var out []Comment
	err := c.walkListing(ctx,
		func(after string, count int) string {
			pp := p
			pp.After = after
			pp.Count = count
			return UserListingURL(name, "comments", pp)
		}, pages, p.Limit,
		func(children []thingEnvelope) (int, error) {
			out = append(out, parseCommentsFlat(children)...)
			return len(out), nil
		})
	if err != nil {
		return out, err
	}
	if p.Limit > 0 && len(out) > p.Limit {
		out = out[:p.Limit]
	}
	return out, nil
}

// SearchPosts walks site-wide or subreddit-scoped post search into Posts.
func (c *Client) SearchPosts(ctx context.Context, query, sub string, p ListingParams, pages int, nsfw bool) ([]Post, error) {
	var out []Post
	err := c.walkListing(ctx,
		func(after string, count int) string {
			pp := p
			pp.After = after
			pp.Count = count
			return SearchURL(query, sub, pp, "link", nsfw)
		}, pages, p.Limit,
		func(children []thingEnvelope) (int, error) {
			out = append(out, parsePosts(children)...)
			return len(out), nil
		})
	if err != nil {
		return out, err
	}
	return trimPosts(out, p.Limit), nil
}

// SearchSubreddits walks subreddit discovery search into lean rows.
func (c *Client) SearchSubreddits(ctx context.Context, query string, p ListingParams, pages int) ([]SubredditResult, error) {
	var out []SubredditResult
	err := c.walkListing(ctx,
		func(after string, count int) string {
			pp := p
			pp.After = after
			pp.Count = count
			return SubredditsSearchURL(query, pp)
		}, pages, p.Limit,
		func(children []thingEnvelope) (int, error) {
			for _, ch := range children {
				if ch.Kind != string(kindSubreddit) {
					continue
				}
				if s, err := parseSubreddit(ch.Data); err == nil {
					out = append(out, SubredditResult{
						Name: s.Name, Title: s.Title, Subscribers: s.Subscribers,
						PublicDescription: s.PublicDescription, Over18: s.Over18, URL: s.URL,
					})
				}
			}
			return len(out), nil
		})
	if err != nil {
		return out, err
	}
	return positionSubs(out, p.Limit), nil
}

// SearchUsers walks user discovery search into lean rows.
func (c *Client) SearchUsers(ctx context.Context, query string, p ListingParams, pages int) ([]UserResult, error) {
	var out []UserResult
	err := c.walkListing(ctx,
		func(after string, count int) string {
			pp := p
			pp.After = after
			pp.Count = count
			return UsersSearchURL(query, pp)
		}, pages, p.Limit,
		func(children []thingEnvelope) (int, error) {
			for _, ch := range children {
				if ch.Kind != "t2" {
					continue
				}
				if u, err := parseUser(ch.Data); err == nil {
					out = append(out, UserResult{
						Name: u.Name, TotalKarma: u.TotalKarma, CreatedUTC: u.CreatedUTC, URL: u.URL,
					})
				}
			}
			return len(out), nil
		})
	if err != nil {
		return out, err
	}
	return positionUsers(out, p.Limit), nil
}

// Duplicates fetches the other discussions of a link.
func (c *Client) Duplicates(ctx context.Context, id string, limit int) ([]Post, error) {
	body, err := c.fetchJSON(ctx, DuplicatesURL(id))
	if err != nil {
		return nil, err
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(body, &arr); err != nil {
		return nil, fmt.Errorf("decode duplicates array: %w", err)
	}
	var out []Post
	if len(arr) >= 2 {
		if ld, err := decodeListing(arr[1]); err == nil {
			out = parsePosts(ld.Children)
		}
	}
	return trimPosts(out, limit), nil
}

// Rules fetches a subreddit's posted rules.
func (c *Client) Rules(ctx context.Context, sub string) ([]Rule, error) {
	body, err := c.fetchJSON(ctx, RulesURL(sub))
	if err != nil {
		return nil, err
	}
	var payload struct {
		Rules []struct {
			Kind            string  `json:"kind"`
			ShortName       string  `json:"short_name"`
			Description     string  `json:"description"`
			ViolationReason string  `json:"violation_reason"`
			Priority        int     `json:"priority"`
			CreatedUTC      float64 `json:"created_utc"`
		} `json:"rules"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	name := cleanName(sub)
	out := make([]Rule, 0, len(payload.Rules))
	for _, r := range payload.Rules {
		out = append(out, Rule{
			Subreddit: name, Kind: r.Kind, ShortName: r.ShortName,
			Description: r.Description, ViolationReason: r.ViolationReason,
			Priority: r.Priority, CreatedUTC: epoch(r.CreatedUTC),
		})
	}
	return out, nil
}

// Moderators fetches a subreddit's moderators.
func (c *Client) Moderators(ctx context.Context, sub string) ([]Moderator, error) {
	body, err := c.fetchJSON(ctx, ModeratorsURL(sub))
	if err != nil {
		return nil, err
	}
	var payload struct {
		Data struct {
			Children []struct {
				Name           string   `json:"name"`
				ID             string   `json:"id"`
				ModPermissions []string `json:"mod_permissions"`
				Date           float64  `json:"date"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	name := cleanName(sub)
	out := make([]Moderator, 0, len(payload.Data.Children))
	for _, m := range payload.Data.Children {
		out = append(out, Moderator{
			Subreddit: name, Name: m.Name, AuthorFullname: m.ID,
			ModPermissions: m.ModPermissions, Date: epoch(m.Date),
		})
	}
	return out, nil
}

// Wiki fetches a subreddit wiki page.
func (c *Client) Wiki(ctx context.Context, sub, page string) (*WikiPage, error) {
	body, err := c.fetchJSON(ctx, WikiURL(sub, page))
	if err != nil {
		return nil, err
	}
	var env thingEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, err
	}
	var d struct {
		ContentMD    string  `json:"content_md"`
		MayRevise    bool    `json:"may_revise"`
		RevisionDate float64 `json:"revision_date"`
		RevisionBy   struct {
			Data struct {
				Name string `json:"name"`
			} `json:"data"`
		} `json:"revision_by"`
	}
	if err := json.Unmarshal(env.Data, &d); err != nil {
		return nil, err
	}
	if page == "" {
		page = "index"
	}
	return &WikiPage{
		Subreddit: cleanName(sub), Page: page, ContentMD: d.ContentMD,
		RevisionBy: d.RevisionBy.Data.Name, RevisionDate: epoch(d.RevisionDate),
		MayRevise: d.MayRevise, URL: BaseURL + "/r/" + cleanName(sub) + "/wiki/" + page,
	}, nil
}

// WikiPages fetches the wiki page index.
func (c *Client) WikiPages(ctx context.Context, sub string) ([]WikiIndexEntry, error) {
	body, err := c.fetchJSON(ctx, WikiPagesURL(sub))
	if err != nil {
		return nil, err
	}
	var env struct {
		Data []string `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, err
	}
	name := cleanName(sub)
	out := make([]WikiIndexEntry, 0, len(env.Data))
	for _, p := range env.Data {
		out = append(out, WikiIndexEntry{Subreddit: name, Page: p, URL: BaseURL + "/r/" + name + "/wiki/" + p})
	}
	return out, nil
}

func trimPosts(p []Post, limit int) []Post {
	if limit > 0 && len(p) > limit {
		return p[:limit]
	}
	return p
}

func positionSubs(s []SubredditResult, limit int) []SubredditResult {
	if limit > 0 && len(s) > limit {
		s = s[:limit]
	}
	for i := range s {
		s[i].Position = i + 1
	}
	return s
}

func positionUsers(u []UserResult, limit int) []UserResult {
	if limit > 0 && len(u) > limit {
		u = u[:limit]
	}
	for i := range u {
		u[i].Position = i + 1
	}
	return u
}
