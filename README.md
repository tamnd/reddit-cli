# reddit

A fast, friendly command line for public [Reddit](https://www.reddit.com) data.
One binary that turns subreddit listings, posts, comment trees, user and
subreddit profiles, search, rules, moderators, wiki pages, and the duplicate
discussions of a link into rich, structured records as a table, JSON, JSONL,
CSV, TSV, or plain URLs.

```
reddit posts golang --sort top --time week -n 5
```

```
SCORE  NUM_COMMENTS  TITLE                                          PERMALINK
1842   312           Go 1.24 is released                            https://www.reddit.com/r/golang/comments/...
1190   208           What are you working on this week?             https://www.reddit.com/r/golang/comments/...
absc   97            A deep dive into the scheduler                 https://www.reddit.com/r/golang/comments/...
640    155           Generics, two years on                         https://www.reddit.com/r/golang/comments/...
512    88            Show: a pure-Go SQLite driver benchmark        https://www.reddit.com/r/golang/comments/...
```

Full documentation: [reddit-cli.tamnd.com](https://reddit-cli.tamnd.com).

> reddit is an independent, open-source tool. It is not affiliated with,
> endorsed by, or sponsored by Reddit, Inc. It reads only public pages, at a
> polite default rate.

## Why

Reddit has an official API, but using it for read-only work means registering an
app, holding a token, and living inside per-app quotas. There is a simpler door:
every public Reddit page also serves a `.json` view of the same content. reddit
is built on that door. It reads the public JSON, classifies what it found
through Reddit's "thing" types (posts, comments, accounts, subreddits), and
hands you records with real fields.

It speaks to `www.reddit.com` over plain HTTPS, with no API key and no account.
The binary is pure Go with no runtime dependencies.

## Install

```sh
go install github.com/tamnd/reddit-cli/cmd/reddit@latest
```

Or grab a prebuilt binary from the [releases page](https://github.com/tamnd/reddit-cli/releases),
install a Linux package (`deb`, `rpm`, `apk`), or pull the container image:

```sh
docker run --rm ghcr.io/tamnd/reddit posts golang
```

Homebrew and Scoop:

```sh
brew install --cask tamnd/tap/reddit
scoop install reddit
```

Build from source:

```sh
git clone https://github.com/tamnd/reddit-cli
cd reddit-cli
make build      # produces ./bin/reddit
```

## Quick start

```sh
reddit posts golang                            # a subreddit's hot listing
reddit posts golang --sort top --time week     # top of the week
reddit comments 1abc23 --sort top -n 200       # a flattened comment tree
reddit user spez                               # a user profile
reddit subreddit golang -o json                # a community as JSON
reddit search "context deadline" --in golang   # search within a subreddit
```

## How it works

Append `.json` to almost any Reddit URL and you get the same content as
structured data. reddit knows the shape and pagination of each of these
endpoints (listings, comment pages, about pages, search, rules, moderators,
wiki, duplicates) and walks the right one from a name or a URL. A listing is
paged through its `after` cursor; a nested comment tree is flattened into one
record per comment, each carrying its depth and parent so the shape survives.
Responses are cached on disk (content-addressed and gzipped) so a repeat call is
instant and does not hit the network.

reddit is polite by default: a two second gap between requests, two workers, and
a descriptive User-Agent, because Reddit rate-limits aggressive and generic
clients the hardest. When Reddit answers with a rate-limit page, a `403`, or its
"whoa there, pardner" interstitial instead of the content, reddit exits cleanly
with code 5 and suggests slowing down or lending a signed-in session with
`--cookies` (a Netscape `cookies.txt` jar exported from your browser). Datacenter
and shared IPs are blocked the hardest; a normal connection at the default rate
rarely sees this.

## Commands

| Command | What it does |
| --- | --- |
| `posts <subreddit>...` | List a subreddit's posts (`--sort`, `--time`) |
| `post <id\|url>...` | Fetch one or more posts (the link, without comments) |
| `comments <id\|url>` | A flattened comment tree (`--sort`, `--depth`, `--expand`) |
| `subreddit <name\|url>...` | Fetch subreddit profiles (alias `sub`, `r`) |
| `user <name\|url>...` | Fetch user profiles (alias `u`) |
| `user-posts <name\|url>` | List a user's submitted posts |
| `user-comments <name\|url>` | List a user's comments |
| `search <query>` | Search posts (`--in`, `--sort`, `--time`, `--nsfw`) |
| `subreddits <query>` | Discover subreddits by name |
| `users <query>` | Discover users by name |
| `rules <subreddit>` | A subreddit's posted rules |
| `mods <subreddit>` | A subreddit's moderators |
| `wiki <subreddit> [page]` | A subreddit wiki page (default: index) |
| `wiki-pages <subreddit>` | A subreddit's wiki page index |
| `duplicates <id\|url>` | The other discussions of a link (alias `dupes`) |
| `id <url\|id>...` | Classify into (kind, id) without fetching |
| `seed <subreddit>...` | Emit post URLs from listings (`--enqueue`) |
| `crawl` | Process the crawl queue (`--max`, `--parse`) |
| `db` | Inspect and export the local store (`info`, `count`, `get`, `export`, `vacuum`) |
| `cache` | Inspect and clear the page cache (`info`, `clear`, `path`) |
| `open <id\|url>` | Open a Reddit page in the default browser |
| `info` | Show configuration, paths, and the disclaimer |
| `version` | Print version, commit, and build date |

## Output

Output is a table on a terminal and JSONL when piped, so it drops straight into
a pipeline. Pick any format explicitly with `-o`:

```sh
reddit posts golang -o json                          # pretty JSON array
reddit posts golang -o jsonl                          # one JSON object per line
reddit user spez -o csv                               # CSV with a header row
reddit posts golang -o url                            # just the URLs
reddit posts golang --fields title,score,num_comments -o tsv
reddit comments 1abc23 --template '{{.author}}: {{.body}}'
```

Choose columns with `--fields`, drop the header with `--no-header`, and apply a
Go `text/template` per record with `--template`.

## Bulk crawling

For dataset-scale work, seed the queue from listings, crawl it, and export the
results from the local SQLite store:

```sh
reddit seed golang --sort top --time month -n 200 --enqueue   # fill the queue
reddit crawl --parse                                          # fetch, cache, parse
reddit db count                                               # how many records
reddit db export --type post --out golang.jsonl               # dump them
```

The crawler is polite by default (2 workers, a 2s spacing) and you can tune it
with `--workers` and `--delay`.

## Exit codes

| Code | Meaning |
| --- | --- |
| 0 | success |
| 1 | error |
| 2 | usage error |
| 3 | no data (not found, empty result) |
| 4 | partial (some items in a batch failed) |
| 5 | blocked (rate-limited or a block page; slow down or try `--cookies`) |

## Configuration

State lives under `$XDG_DATA_HOME/reddit` (or `~/.local/share/reddit`),
overridable with `--data-dir` or `REDDIT_DATA_DIR`. The page cache and the
SQLite store both sit there. Politeness and networking knobs (`--delay`,
`--workers`, `--timeout`, `--retries`, `--cache-ttl`, `--no-cache`, `--refresh`,
`--user-agent`, `--cookies`) are global flags on every command. Run `reddit
info` to see the resolved paths and `reddit <command> --help` for the full
surface.

## Library

The fetching and parsing live in the `reddit` package, so you can read Reddit
from your own program without the CLI:

```go
import "github.com/tamnd/reddit-cli/reddit"

c := reddit.NewClient(reddit.DefaultConfig())
posts, err := c.Posts(ctx, "golang", reddit.ListingParams{Sort: "top", Limit: 25}, 1)
```

## Development

```sh
make build      # build ./bin/reddit
make test       # go test ./...
make vet        # go vet ./...
make fmt        # gofmt -s -w .
```

## License

[Apache-2.0](LICENSE).
