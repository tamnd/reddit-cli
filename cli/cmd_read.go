package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tamnd/reddit-cli/reddit"
)

// posts ───────────────────────────────────────────────────────────────────────

func (a *App) postsCmd() *cobra.Command {
	var sort, window string
	cmd := &cobra.Command{
		Use:   "posts <subreddit> [subreddit ...]",
		Short: "List a subreddit's posts by sort",
		Args:  cobra.MinimumNArgs(1),
		Example: "  reddit posts golang\n" +
			"  reddit posts golang --sort top --time week -n 25\n" +
			"  reddit posts r/golang r/rust -o json",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !reddit.ValidSort(sort) {
				return codeError(exitUsage, fmt.Errorf("invalid --sort %q (hot|new|top|rising|controversial)", sort))
			}
			if window != "" && !reddit.ValidTime(window) {
				return codeError(exitUsage, fmt.Errorf("invalid --time %q (hour|day|week|month|year|all)", window))
			}
			ctx := cmd.Context()
			var posts []reddit.Post
			for _, arg := range args {
				p, err := a.client.Posts(ctx, arg, a.listingParams(sort, window), a.walkPages())
				if err != nil {
					if len(args) == 1 {
						return mapFetchErr(err)
					}
					a.progressf("posts %s: %v", arg, err)
					continue
				}
				a.storePosts(p)
				posts = append(posts, p...)
			}
			return a.renderOrEmpty(posts, len(posts))
		},
	}
	cmd.Flags().StringVar(&sort, "sort", "hot", "sort: hot|new|top|rising|controversial")
	cmd.Flags().StringVar(&window, "time", "", "time window for top/controversial: hour|day|week|month|year|all")
	return cmd
}

// post ─────────────────────────────────────────────────────────────────────────

func (a *App) postCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "post <id|url> [id|url ...]",
		Short:   "Fetch one or more posts (the link, without comments)",
		Args:    cobra.MinimumNArgs(1),
		Example: "  reddit post 1abc23\n  reddit post https://www.reddit.com/r/golang/comments/1abc23/title/",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var posts []*reddit.Post
			for _, arg := range args {
				p, err := a.client.Post(ctx, arg)
				if err != nil {
					if len(args) == 1 {
						return mapFetchErr(err)
					}
					a.progressf("post %s: %v", arg, err)
					continue
				}
				if a.store != nil {
					_ = a.store.Put("post", p.PostID, p.Permalink, p)
				}
				posts = append(posts, p)
			}
			return a.renderOrEmpty(posts, len(posts))
		},
	}
}

// comments ─────────────────────────────────────────────────────────────────────

func (a *App) commentsCmd() *cobra.Command {
	var sort string
	var depth int
	var expand bool
	cmd := &cobra.Command{
		Use:   "comments <id|url>",
		Short: "Fetch a post's comment tree, flattened into records",
		Args:  cobra.ExactArgs(1),
		Long: "comments returns one record per comment, carrying depth and parent links\n" +
			"so a flattened tree keeps its shape. Use --expand to follow the collapsed\n" +
			"\"load more\" stubs through the morechildren endpoint.",
		Example: "  reddit comments 1abc23 --sort top -n 200\n  reddit comments 1abc23 --expand",
		RunE: func(cmd *cobra.Command, args []string) error {
			comments, err := a.client.Comments(cmd.Context(), args[0], sort, a.limit, depth, expand)
			if err != nil {
				return mapFetchErr(err)
			}
			if a.store != nil {
				for i := range comments {
					_ = a.store.Put("comment", comments[i].CommentID, comments[i].Permalink, comments[i])
				}
			}
			return a.renderOrEmpty(comments, len(comments))
		},
	}
	cmd.Flags().StringVar(&sort, "sort", "confidence", "comment sort: confidence|top|new|controversial|old|qa")
	cmd.Flags().IntVar(&depth, "depth", 0, "maximum tree depth to request (0 = server default)")
	cmd.Flags().BoolVar(&expand, "expand", false, "expand collapsed \"load more\" stubs via morechildren")
	return cmd
}

// subreddit ─────────────────────────────────────────────────────────────────────

func (a *App) subredditCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "subreddit <name|url> [name|url ...]",
		Aliases: []string{"sub", "r"},
		Short:   "Fetch one or more subreddit profiles",
		Args:    cobra.MinimumNArgs(1),
		Example: "  reddit subreddit golang\n  reddit subreddit r/golang r/rust -o csv",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var subs []*reddit.Subreddit
			for _, arg := range args {
				s, err := a.client.Subreddit(ctx, arg)
				if err != nil {
					if len(args) == 1 {
						return mapFetchErr(err)
					}
					a.progressf("subreddit %s: %v", arg, err)
					continue
				}
				if a.store != nil {
					_ = a.store.Put("subreddit", s.Name, s.URL, s)
				}
				subs = append(subs, s)
			}
			return a.renderOrEmpty(subs, len(subs))
		},
	}
}

// user ─────────────────────────────────────────────────────────────────────────

func (a *App) userCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "user <name|url> [name|url ...]",
		Aliases: []string{"u"},
		Short:   "Fetch one or more user profiles",
		Args:    cobra.MinimumNArgs(1),
		Example: "  reddit user spez\n  reddit user u/spez -o json",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var users []*reddit.User
			for _, arg := range args {
				u, err := a.client.User(ctx, arg)
				if err != nil {
					if len(args) == 1 {
						return mapFetchErr(err)
					}
					a.progressf("user %s: %v", arg, err)
					continue
				}
				if a.store != nil {
					_ = a.store.Put("user", u.Name, u.URL, u)
				}
				users = append(users, u)
			}
			return a.renderOrEmpty(users, len(users))
		},
	}
}

// user-posts ─────────────────────────────────────────────────────────────────────

func (a *App) userPostsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "user-posts <name|url>",
		Short:   "List a user's submitted posts",
		Args:    cobra.ExactArgs(1),
		Example: "  reddit user-posts spez -n 50",
		RunE: func(cmd *cobra.Command, args []string) error {
			posts, err := a.client.UserPosts(cmd.Context(), args[0], a.listingParams("new", ""), a.walkPages())
			if err != nil {
				return mapFetchErr(err)
			}
			a.storePosts(posts)
			return a.renderOrEmpty(posts, len(posts))
		},
	}
}

// user-comments ───────────────────────────────────────────────────────────────────

func (a *App) userCommentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "user-comments <name|url>",
		Short:   "List a user's comments",
		Args:    cobra.ExactArgs(1),
		Example: "  reddit user-comments spez -n 50",
		RunE: func(cmd *cobra.Command, args []string) error {
			comments, err := a.client.UserComments(cmd.Context(), args[0], a.listingParams("new", ""), a.walkPages())
			if err != nil {
				return mapFetchErr(err)
			}
			if a.store != nil {
				for i := range comments {
					_ = a.store.Put("comment", comments[i].CommentID, comments[i].Permalink, comments[i])
				}
			}
			return a.renderOrEmpty(comments, len(comments))
		},
	}
}

// search ─────────────────────────────────────────────────────────────────────────

func (a *App) searchCmd() *cobra.Command {
	var in, sort, window string
	var nsfw bool
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search posts site-wide or within a subreddit",
		Args:  cobra.MinimumNArgs(1),
		Example: "  reddit search \"generics\" --in golang -n 25\n" +
			"  reddit search \"context deadline\" --sort new",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := a.listingParams(sort, window)
			posts, err := a.client.SearchPosts(cmd.Context(), strings.Join(args, " "), in, p, a.walkPages(), nsfw)
			if err != nil {
				return mapFetchErr(err)
			}
			a.storePosts(posts)
			return a.renderOrEmpty(posts, len(posts))
		},
	}
	cmd.Flags().StringVar(&in, "in", "", "restrict the search to this subreddit")
	cmd.Flags().StringVar(&sort, "sort", "relevance", "sort: relevance|hot|top|new|comments")
	cmd.Flags().StringVar(&window, "time", "", "time window: hour|day|week|month|year|all")
	cmd.Flags().BoolVar(&nsfw, "nsfw", false, "include over-18 results")
	return cmd
}

// subreddits ─────────────────────────────────────────────────────────────────────

func (a *App) subredditsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "subreddits <query>",
		Short:   "Discover subreddits matching a query",
		Args:    cobra.MinimumNArgs(1),
		Example: "  reddit subreddits programming -n 20",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := a.client.SearchSubreddits(cmd.Context(), strings.Join(args, " "), a.listingParams("", ""), a.walkPages())
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(res, len(res))
		},
	}
}

// users ─────────────────────────────────────────────────────────────────────────

func (a *App) usersCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "users <query>",
		Short:   "Discover users matching a query",
		Args:    cobra.MinimumNArgs(1),
		Example: "  reddit users gallowboob -n 20",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := a.client.SearchUsers(cmd.Context(), strings.Join(args, " "), a.listingParams("", ""), a.walkPages())
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(res, len(res))
		},
	}
}

// rules ─────────────────────────────────────────────────────────────────────────

func (a *App) rulesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rules <subreddit>",
		Short:   "List a subreddit's posted rules",
		Args:    cobra.ExactArgs(1),
		Example: "  reddit rules golang",
		RunE: func(cmd *cobra.Command, args []string) error {
			rules, err := a.client.Rules(cmd.Context(), args[0])
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(rules, len(rules))
		},
	}
}

// mods ─────────────────────────────────────────────────────────────────────────

func (a *App) modsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "mods <subreddit>",
		Short:   "List a subreddit's moderators",
		Args:    cobra.ExactArgs(1),
		Example: "  reddit mods golang",
		RunE: func(cmd *cobra.Command, args []string) error {
			mods, err := a.client.Moderators(cmd.Context(), args[0])
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(mods, len(mods))
		},
	}
}

// wiki ─────────────────────────────────────────────────────────────────────────

func (a *App) wikiCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "wiki <subreddit> [page]",
		Short:   "Fetch a subreddit wiki page (default: index)",
		Args:    cobra.RangeArgs(1, 2),
		Example: "  reddit wiki golang\n  reddit wiki golang faq",
		RunE: func(cmd *cobra.Command, args []string) error {
			page := ""
			if len(args) == 2 {
				page = args[1]
			}
			w, err := a.client.Wiki(cmd.Context(), args[0], page)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty([]*reddit.WikiPage{w}, 1)
		},
	}
}

// wiki-pages ─────────────────────────────────────────────────────────────────────

func (a *App) wikiPagesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "wiki-pages <subreddit>",
		Short:   "List a subreddit's wiki page index",
		Args:    cobra.ExactArgs(1),
		Example: "  reddit wiki-pages golang",
		RunE: func(cmd *cobra.Command, args []string) error {
			pages, err := a.client.WikiPages(cmd.Context(), args[0])
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(pages, len(pages))
		},
	}
}

// duplicates ─────────────────────────────────────────────────────────────────────

func (a *App) duplicatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "duplicates <id|url>",
		Aliases: []string{"dupes"},
		Short:   "List the other discussions of a link",
		Args:    cobra.ExactArgs(1),
		Example: "  reddit duplicates 1abc23 -n 20",
		RunE: func(cmd *cobra.Command, args []string) error {
			posts, err := a.client.Duplicates(cmd.Context(), args[0], a.limit)
			if err != nil {
				return mapFetchErr(err)
			}
			a.storePosts(posts)
			return a.renderOrEmpty(posts, len(posts))
		},
	}
}

// id ─────────────────────────────────────────────────────────────────────────────

func (a *App) idCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "id <url|id> [url|id ...]",
		Short: "Classify a URL or id into (kind, id) without fetching",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			type row struct {
				Input string `json:"input"`
				Kind  string `json:"kind"`
				ID    string `json:"id"`
			}
			rows := make([]row, 0, len(args))
			for _, arg := range args {
				k, id := reddit.Classify(arg)
				rows = append(rows, row{Input: arg, Kind: k, ID: id})
			}
			return a.renderOrEmpty(rows, len(rows))
		},
	}
}

// storePosts upserts a batch of posts when a store is open.
func (a *App) storePosts(posts []reddit.Post) {
	if a.store == nil {
		return
	}
	for i := range posts {
		_ = a.store.Put("post", posts[i].PostID, posts[i].Permalink, posts[i])
	}
}
