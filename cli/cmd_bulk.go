package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/tamnd/reddit-cli/reddit"
	"golang.org/x/sync/errgroup"
)

// seed ──────────────────────────────────────────────────────────────────────

func (a *App) seedCmd() *cobra.Command {
	var (
		sort    string
		window  string
		enqueue bool
	)
	cmd := &cobra.Command{
		Use:   "seed <subreddit> [subreddit ...]",
		Short: "Walk subreddit listings and emit post URLs to crawl",
		Long: "seed walks one or more subreddit listings and emits the comment-page URL\n" +
			"of every post it finds. Add --enqueue to load those URLs into the crawl\n" +
			"queue, then run crawl to fetch and parse them.",
		Args: cobra.MinimumNArgs(1),
		Example: "  reddit seed golang --sort top --time week\n" +
			"  reddit seed golang rust --enqueue",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !reddit.ValidSort(sort) {
				return codeError(exitUsage, fmt.Errorf("invalid --sort %q (hot|new|top|rising|controversial)", sort))
			}
			ctx := cmd.Context()
			type row struct {
				URL        string `json:"url"`
				EntityType string `json:"entity_type"`
			}
			var rows []row
			var st *reddit.Store
			var err error
			if enqueue {
				if st, err = a.openStore(); err != nil {
					return codeError(exitError, err)
				}
			}
			for _, sub := range args {
				posts, perr := a.client.Posts(ctx, sub, a.listingParams(sort, window), a.walkPages())
				if perr != nil {
					if len(args) == 1 {
						return mapFetchErr(perr)
					}
					a.progressf("seed %s: %v", sub, perr)
					continue
				}
				for i := range posts {
					u := reddit.PostURL(posts[i].PostID)
					rows = append(rows, row{URL: u, EntityType: "post"})
					if enqueue {
						_ = st.Enqueue(ctx, u, "post", 0)
					}
				}
				a.progressf("seed %s: %d posts", sub, len(posts))
			}
			if enqueue {
				a.progressf("enqueued %d URLs into %s", len(rows), a.storePath())
			}
			return a.renderOrEmpty(rows, len(rows))
		},
	}
	cmd.Flags().StringVar(&sort, "sort", "hot", "sort: hot|new|top|rising|controversial")
	cmd.Flags().StringVar(&window, "time", "", "time window for top/controversial")
	cmd.Flags().BoolVar(&enqueue, "enqueue", false, "enqueue discovered URLs into the crawl queue")
	return cmd
}

// crawl ─────────────────────────────────────────────────────────────────────

func (a *App) crawlCmd() *cobra.Command {
	var (
		maxItems int
		parse    bool
	)
	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Process the crawl queue (fetch, cache, optionally parse)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			st, err := a.openStore()
			if err != nil {
				return codeError(exitError, err)
			}
			if err := st.ResetActive(); err != nil {
				return codeError(exitError, err)
			}
			var (
				processed int
				failed    int
				mu        sync.Mutex
			)
			for maxItems <= 0 || processed < maxItems {
				batch := a.cfg.Workers * 2
				if maxItems > 0 && processed+batch > maxItems {
					batch = maxItems - processed
				}
				items, err := st.NextPending(ctx, batch)
				if err != nil {
					return codeError(exitError, err)
				}
				if len(items) == 0 {
					break
				}
				g, gctx := errgroup.WithContext(ctx)
				g.SetLimit(a.cfg.Workers)
				for _, it := range items {
					g.Go(func() error {
						ferr := a.crawlOne(gctx, st, it, parse)
						mu.Lock()
						defer mu.Unlock()
						if ferr != nil {
							failed++
							_ = st.MarkFailed(gctx, it.ID)
							a.progressf("crawl %s: %v", it.URL, ferr)
							return nil
						}
						processed++
						_ = st.MarkFetched(gctx, it.ID, "")
						return nil
					})
				}
				if err := g.Wait(); err != nil {
					return codeError(exitError, err)
				}
				a.progressf("crawled %d (failed %d)", processed, failed)
			}
			stats, _ := st.QueueStats()
			fmt.Printf("done: processed=%d failed=%d queue=%v\n", processed, failed, stats)
			if processed == 0 {
				return codeError(exitNoData, nil)
			}
			if failed > 0 {
				return codeError(exitPartial, nil)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&maxItems, "max", 0, "process at most this many items (0 = drain queue)")
	cmd.Flags().BoolVar(&parse, "parse", false, "parse each fetched post and its comments into the records table")
	return cmd
}

// crawlOne fetches a queued post URL and, when parse is set, stores the post and
// its comments.
func (a *App) crawlOne(ctx context.Context, st *reddit.Store, it reddit.QueueItem, parse bool) error {
	_, id := reddit.Classify(it.URL)
	if id == "" {
		id = it.URL
	}
	post, err := a.client.Post(ctx, id)
	if err != nil {
		return err
	}
	if !parse {
		return nil
	}
	_ = st.Put("post", post.PostID, post.Permalink, post)
	comments, err := a.client.Comments(ctx, id, "confidence", 0, 0, false)
	if err != nil {
		return err
	}
	for i := range comments {
		_ = st.Put("comment", comments[i].CommentID, comments[i].Permalink, comments[i])
	}
	return nil
}

// db ──────────────────────────────────────────────────────────────────────────

func (a *App) dbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Inspect and export the local store",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(a.dbInfoCmd(), a.dbCountCmd(), a.dbGetCmd(), a.dbExportCmd(), a.dbVacuumCmd())
	return cmd
}

func (a *App) dbInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Summarize stored records and the crawl queue",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			st, err := a.openStore()
			if err != nil {
				return codeError(exitError, err)
			}
			counts, err := st.CountsByType()
			if err != nil {
				return codeError(exitError, err)
			}
			q, _ := st.QueueStats()
			type row struct {
				EntityType string `json:"entity_type"`
				Count      int    `json:"count"`
			}
			var rows []row
			for k, v := range counts {
				rows = append(rows, row{EntityType: k, Count: v})
			}
			if len(rows) == 0 {
				a.progressf("store has no records yet (%s)", a.storePath())
			}
			a.progressf("queue: %v", q)
			return a.render(rows)
		},
	}
}

func (a *App) dbCountCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "count [entity-type]",
		Short: "Count stored records (all types, or one)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			st, err := a.openStore()
			if err != nil {
				return codeError(exitError, err)
			}
			t := ""
			if len(args) == 1 {
				t = args[0]
			}
			n, err := st.Count(t)
			if err != nil {
				return codeError(exitError, err)
			}
			fmt.Println(n)
			return nil
		},
	}
}

func (a *App) dbGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <entity-type> <id>",
		Short: "Print a stored record as JSON",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			st, err := a.openStore()
			if err != nil {
				return codeError(exitError, err)
			}
			data, err := st.Get(args[0], args[1])
			if err != nil {
				return mapFetchErr(err)
			}
			var buf any
			_ = json.Unmarshal(data, &buf)
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(buf)
		},
	}
}

func (a *App) dbExportCmd() *cobra.Command {
	var (
		entityType string
		out        string
	)
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export stored records to JSONL",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			st, err := a.openStore()
			if err != nil {
				return codeError(exitError, err)
			}
			w := os.Stdout
			if out != "" {
				f, err := os.Create(out)
				if err != nil {
					return codeError(exitError, err)
				}
				defer func() { _ = f.Close() }()
				w = f
			}
			n := 0
			err = st.Each(entityType, func(_ string, data []byte) error {
				n++
				_, werr := fmt.Fprintln(w, string(data))
				return werr
			})
			if err != nil {
				return codeError(exitError, err)
			}
			a.progressf("exported %d records", n)
			if n == 0 {
				return codeError(exitNoData, nil)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&entityType, "type", "", "only export this entity type")
	cmd.Flags().StringVar(&out, "out", "", "output file (default: stdout)")
	return cmd
}

func (a *App) dbVacuumCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vacuum",
		Short: "Compact the store database file",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			st, err := a.openStore()
			if err != nil {
				return codeError(exitError, err)
			}
			if err := st.Vacuum(); err != nil {
				return codeError(exitError, err)
			}
			fmt.Println("ok")
			return nil
		},
	}
}

func (a *App) storePath() string {
	if a.storePtr != "" {
		return a.storePtr
	}
	return a.cfg.StorePath()
}

// open ──────────────────────────────────────────────────────────────────────

func (a *App) openCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <id|url>",
		Short: "Open a Reddit page in the default browser",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			url := reddit.ResolveURL(args[0])
			if url == "" {
				return codeError(exitUsage, fmt.Errorf("could not resolve %q to a URL", args[0]))
			}
			return openBrowser(url)
		},
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// cache ─────────────────────────────────────────────────────────────────────

func (a *App) cacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Inspect and clear the on-disk response cache",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "info",
			Short: "Show cache location, file count, and size",
			Args:  cobra.NoArgs,
			RunE: func(_ *cobra.Command, _ []string) error {
				files, total, err := a.cache.Stats()
				if err != nil {
					return codeError(exitError, err)
				}
				fmt.Printf("dir:   %s\nfiles: %d\nsize:  %s\n", a.cfg.CacheDir(), files, humanize.Bytes(uint64(total)))
				return nil
			},
		},
		&cobra.Command{
			Use:   "clear",
			Short: "Delete the entire response cache",
			Args:  cobra.NoArgs,
			RunE: func(_ *cobra.Command, _ []string) error {
				if err := a.cache.Clear(); err != nil {
					return codeError(exitError, err)
				}
				fmt.Println("cache cleared")
				return nil
			},
		},
		&cobra.Command{
			Use:   "path <url>",
			Short: "Print the cache path for a URL",
			Args:  cobra.ExactArgs(1),
			RunE: func(_ *cobra.Command, args []string) error {
				fmt.Println(a.cache.Path(args[0]))
				return nil
			},
		},
	)
	return cmd
}

// info ──────────────────────────────────────────────────────────────────────

func (a *App) infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show configuration, paths, and the affiliation disclaimer",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf(`reddit %s

reddit is an independent, open-source tool. It is not affiliated with,
endorsed by, or sponsored by Reddit, Inc. It reads only public pages, at a
polite default rate, and respects each community's privacy and block pages.

data dir:   %s
cache dir:  %s
store:      %s
workers:    %d
delay:      %s
cache TTL:  %s
user-agent: %s
`,
				Version, a.cfg.DataDir, a.cfg.CacheDir(), a.storePath(),
				a.cfg.Workers, a.cfg.Delay, a.cfg.CacheTTL, a.cfg.UserAgent)
			return nil
		},
	}
}

// version ─────────────────────────────────────────────────────────────────────

func (a *App) versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version, commit, and build date",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("reddit %s (commit %s, built %s)\n", Version, Commit, Date)
			return nil
		},
	}
}
