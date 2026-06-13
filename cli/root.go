package cli

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/tamnd/reddit-cli/reddit"
)

// Build metadata, injected via -ldflags by the Makefile/goreleaser.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// App holds shared state threaded through every command.
type App struct {
	cfg      reddit.Config
	client   *reddit.Client
	cache    *reddit.Cache
	store    *reddit.Store
	storePtr string

	// global flags
	output   string
	fields   []string
	noHeader bool
	template string
	color    string
	limit    int
	pages    int
	quiet    bool
}

// exit codes (see spec §6).
const (
	exitError   = 1
	exitUsage   = 2
	exitNoData  = 3
	exitPartial = 4
	exitBlocked = 5
)

// ExitError carries a process exit code up to main.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit %d", e.Code)
}

func (e *ExitError) Unwrap() error { return e.Err }

func codeError(code int, err error) error { return &ExitError{Code: code, Err: err} }

// NewRootCmd builds the full command tree.
func NewRootCmd() *cobra.Command {
	app := &App{cfg: reddit.DefaultConfig()}

	root := &cobra.Command{
		Use:   "reddit",
		Short: "Read public Reddit data as structured records",
		Long: "reddit reads the public .json view that every Reddit path exposes:\n" +
			"subreddit listings, posts, comment trees, user and subreddit profiles,\n" +
			"search, rules, moderators, wiki pages, and the duplicate discussions of a\n" +
			"link. It returns rich records as table, JSON, JSONL, CSV, TSV, or URLs.\n\n" +
			"reddit is an independent tool and is not affiliated with, endorsed by, or\n" +
			"sponsored by Reddit, Inc.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return app.setup(cmd)
		},
		PersistentPostRun: func(_ *cobra.Command, _ []string) {
			if app.store != nil {
				_ = app.store.Close()
			}
		},
	}

	pf := root.PersistentFlags()
	pf.StringVarP(&app.output, "output", "o", "auto", "output: table|json|jsonl|csv|tsv|url|raw (auto = table on a TTY, jsonl when piped)")
	pf.StringSliceVar(&app.fields, "fields", nil, "comma-separated columns to include")
	pf.BoolVar(&app.noHeader, "no-header", false, "omit the header row in table/csv/tsv")
	pf.StringVar(&app.template, "template", "", "Go text/template applied per record")
	pf.StringVar(&app.color, "color", "auto", "color: auto|always|never")
	pf.IntVarP(&app.limit, "limit", "n", 0, "limit number of records (0 = no limit)")
	pf.IntVar(&app.pages, "pages", 1, "listing pages to walk (0 = until exhausted or limit)")
	pf.BoolVarP(&app.quiet, "quiet", "q", false, "suppress progress on stderr")

	pf.IntVarP(&app.cfg.Workers, "workers", "j", reddit.DefaultWorkers, "concurrent workers for bulk/crawl")
	pf.DurationVar(&app.cfg.Delay, "delay", reddit.DefaultDelay, "minimum spacing between requests")
	pf.DurationVar(&app.cfg.Timeout, "timeout", reddit.DefaultTimeout, "per-request timeout")
	pf.IntVar(&app.cfg.Retries, "retries", reddit.DefaultRetries, "retry attempts on 429/5xx")
	pf.DurationVar(&app.cfg.CacheTTL, "cache-ttl", reddit.DefaultCacheTTL, "on-disk cache freshness window")
	pf.BoolVar(&app.cfg.NoCache, "no-cache", false, "bypass the on-disk response cache")
	pf.BoolVar(&app.cfg.Refresh, "refresh", false, "force re-fetch and overwrite the cache")
	pf.StringVar(&app.cfg.DataDir, "data-dir", app.cfg.DataDir, "root directory for cache and store")
	pf.StringVar(&app.storePtr, "store", "", "SQLite store path (default: <data-dir>/reddit.db)")
	pf.StringVar(&app.cfg.UserAgent, "user-agent", reddit.DefaultUserAgent, "User-Agent sent with each request")
	pf.StringVar(&app.cfg.CookiePath, "cookies", "", "Netscape cookie jar for a lent session")

	root.AddCommand(
		app.postsCmd(),
		app.postCmd(),
		app.commentsCmd(),
		app.subredditCmd(),
		app.userCmd(),
		app.userPostsCmd(),
		app.userCommentsCmd(),
		app.searchCmd(),
		app.subredditsCmd(),
		app.usersCmd(),
		app.rulesCmd(),
		app.modsCmd(),
		app.wikiCmd(),
		app.wikiPagesCmd(),
		app.duplicatesCmd(),
		app.idCmd(),
		app.seedCmd(),
		app.crawlCmd(),
		app.dbCmd(),
		app.openCmd(),
		app.cacheCmd(),
		app.infoCmd(),
		app.versionCmd(),
	)
	return root
}

// setup resolves output defaults and constructs the shared client and cache.
func (a *App) setup(_ *cobra.Command) error {
	if a.output == "" || a.output == "auto" {
		if isatty.IsTerminal(os.Stdout.Fd()) {
			a.output = string(FormatTable)
		} else {
			a.output = string(FormatJSONL)
		}
	}
	if !Format(a.output).Valid() {
		return codeError(exitUsage, fmt.Errorf("unknown output format %q", a.output))
	}
	if a.cfg.CookiePath != "" {
		cookies, err := reddit.LoadCookies(a.cfg.CookiePath)
		if err != nil {
			return codeError(exitUsage, fmt.Errorf("load cookies: %w", err))
		}
		c, err := reddit.NewClientWithCookies(a.cfg, cookies)
		if err != nil {
			return err
		}
		a.client = c
	} else {
		a.client = reddit.NewClient(a.cfg)
	}
	a.cache = reddit.NewCache(a.cfg)
	a.client = a.client.WithCache(a.cache)
	return nil
}

// openStore lazily opens the SQLite store.
func (a *App) openStore() (*reddit.Store, error) {
	if a.store != nil {
		return a.store, nil
	}
	path := a.storePtr
	if path == "" {
		path = a.cfg.StorePath()
	}
	st, err := reddit.OpenStore(path)
	if err != nil {
		return nil, err
	}
	a.store = st
	return st, nil
}

// listingParams builds the shared pagination/windowing knobs from flags.
func (a *App) listingParams(sort, window string) reddit.ListingParams {
	limit := a.limit
	if limit <= 0 || limit > reddit.MaxPageLimit {
		limit = reddit.MaxPageLimit
	}
	return reddit.ListingParams{Sort: sort, Time: window, Limit: limit}
}

// walkPages returns how many listing pages to walk: --pages, widened when a
// --limit larger than one page is asked for and --pages was left at one.
func (a *App) walkPages() int {
	if a.limit > reddit.MaxPageLimit && a.pages <= 1 {
		return 0
	}
	return a.pages
}

// render writes records using the resolved global flags.
func (a *App) render(records any) error {
	r := NewRenderer(os.Stdout, Format(a.output), a.fields, a.noHeader, a.template)
	return r.Render(records)
}

// renderOrEmpty renders records, mapping an empty result to exit code 3.
func (a *App) renderOrEmpty(records any, n int) error {
	if err := a.render(records); err != nil {
		return err
	}
	if n == 0 {
		return codeError(exitNoData, nil)
	}
	return nil
}

// progressf prints a progress line to stderr unless --quiet.
func (a *App) progressf(format string, args ...any) {
	if a.quiet {
		return
	}
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// mapFetchErr converts a library error into the right exit code.
func mapFetchErr(err error) error {
	switch {
	case err == nil:
		return nil
	case isBlocked(err):
		return codeError(exitBlocked, fmt.Errorf("%w\nhint: slow down with --delay, or pass --cookies to lend a signed-in session", err))
	case isNotFound(err):
		return codeError(exitNoData, err)
	default:
		return codeError(exitError, err)
	}
}
