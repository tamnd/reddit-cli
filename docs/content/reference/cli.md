---
title: "CLI"
description: "Every command and subcommand, with the flags that matter and one example each."
weight: 10
---

```
reddit <command> [args] [flags]
```

Run `reddit <command> --help` for the full flag list on any command. This page
is the map. Every command reads the public `.json` view; when Reddit rate-limits
or blocks a request, the command exits with code 5. See
[troubleshooting](/reference/troubleshooting/).

## Commands

| Command | What it does |
|---|---|
| `posts` | List a subreddit's posts by sort and time window |
| `post` | Fetch one or more posts (the link, without comments) |
| `comments` | Fetch a post's comment tree, flattened into records |
| `subreddit` | Fetch one or more subreddit profiles (alias `sub`, `r`) |
| `user` | Fetch one or more user profiles (alias `u`) |
| `user-posts` | List a user's submitted posts |
| `user-comments` | List a user's comments |
| `search` | Search posts site-wide or within a subreddit |
| `subreddits` | Discover subreddits matching a query |
| `users` | Discover users matching a query |
| `rules` | List a subreddit's posted rules |
| `mods` | List a subreddit's moderators |
| `wiki` | Fetch a subreddit wiki page (default: index) |
| `wiki-pages` | List a subreddit's wiki page index |
| `duplicates` | List the other discussions of a link (alias `dupes`) |
| `id` | Classify a URL or id into (kind, id) without fetching |
| `seed` | Walk subreddit listings and emit post URLs to crawl |
| `crawl` | Process the crawl queue: fetch, cache, optionally parse |
| `db` | Inspect and export the local SQLite store |
| `cache` | Inspect and clear the on-disk page cache |
| `open` | Open a Reddit page in the default browser |
| `info` | Show configuration, paths, and the affiliation disclaimer |
| `version` | Print version, commit, and build date |

## posts

```
reddit posts <subreddit> [subreddit ...] [flags]
```

Lists a subreddit's posts. `--sort` takes `hot`, `new`, `top`, `rising`, or
`controversial` (default `hot`); `--time` takes `hour`, `day`, `week`, `month`,
`year`, or `all` for the `top` and `controversial` sorts. Pages with `-n` and
`--pages`.

```bash
reddit posts golang --sort top --time week -n 25
```

## post

```
reddit post <id|url> [id|url ...]
```

Fetches one or more posts, the link itself without its comments.

```bash
reddit post 1abc23 -o json
```

## comments

```
reddit comments <id|url> [flags]
```

Fetches a post's comment tree flattened into one record per comment. `--sort`
takes `confidence` (default), `top`, `new`, `controversial`, `old`, or `qa`.
`--depth N` limits the tree depth. `--expand` follows the collapsed "load more"
stubs.

```bash
reddit comments 1abc23 --sort top -n 200 --expand
```

## subreddit

```
reddit subreddit <name|url> [name|url ...]
```

Fetches one or more subreddit profiles. Aliases: `sub`, `r`.

```bash
reddit subreddit golang rust -o csv
```

## user

```
reddit user <name|url> [name|url ...]
```

Fetches one or more user profiles. Alias: `u`.

```bash
reddit user spez -o json
```

## user-posts

```
reddit user-posts <name|url>
```

Lists a user's submitted posts (newest first).

```bash
reddit user-posts spez -n 50
```

## user-comments

```
reddit user-comments <name|url>
```

Lists a user's comments (newest first).

```bash
reddit user-comments spez -n 50
```

## search

```
reddit search <query> [flags]
```

Searches posts. `--in <subreddit>` restricts to one community. `--sort` takes
`relevance` (default), `hot`, `top`, `new`, or `comments`. `--time` takes the
usual windows. `--nsfw` includes over-18 results.

```bash
reddit search "context deadline" --in golang -n 25
```

## subreddits

```
reddit subreddits <query>
```

Discovers subreddits matching a query.

```bash
reddit subreddits programming -n 20
```

## users

```
reddit users <query>
```

Discovers users matching a query.

```bash
reddit users gallowboob -n 20
```

## rules

```
reddit rules <subreddit>
```

Lists a subreddit's posted rules.

```bash
reddit rules golang
```

## mods

```
reddit mods <subreddit>
```

Lists a subreddit's moderators.

```bash
reddit mods golang
```

## wiki

```
reddit wiki <subreddit> [page]
```

Fetches a subreddit wiki page. With one argument it reads the index; a second
argument names a page.

```bash
reddit wiki golang faq
```

## wiki-pages

```
reddit wiki-pages <subreddit>
```

Lists a subreddit's wiki page index.

```bash
reddit wiki-pages golang
```

## duplicates

```
reddit duplicates <id|url>
```

Lists the other discussions of the same link. Alias: `dupes`.

```bash
reddit duplicates 1abc23 -n 20
```

## id

```
reddit id <url|id> [url|id ...]
```

Classifies each argument into a (kind, id) pair without fetching. Pure local,
never blocked.

```bash
reddit id https://www.reddit.com/r/golang/comments/1abc23/title/
```

## seed

```
reddit seed <subreddit> [subreddit ...] [flags]
```

Walks subreddit listings and emits the comment-page URL of every post. Same
`--sort` and `--time` as `posts`. `--enqueue` loads the URLs into the crawl
queue instead of only printing them.

```bash
reddit seed golang --sort top --time week --enqueue
```

## crawl

```
reddit crawl [flags]
```

Processes the crawl queue: fetch, cache, and with `--parse` store each post and
its comments. `--max N` caps the count (`0` drains the queue). Uses `--workers`
and `--delay`. Exit 3 if nothing processed, exit 4 if some failed.

```bash
reddit crawl --max 50 --parse
```

## db

| Subcommand | Does |
|---|---|
| `db info` | Summarize stored records and the crawl queue |
| `db count [entity-type]` | Count stored records |
| `db get <entity-type> <id>` | Print a stored record as JSON |
| `db export [--type T] [--out file]` | Export stored records to JSONL |
| `db vacuum` | Reclaim space in the store |

```bash
reddit db export --type post --out posts.jsonl
```

## cache

| Subcommand | Does |
|---|---|
| `cache info` | Location, file count, and size |
| `cache clear` | Remove every cached page |
| `cache path <url>` | Print the cache file path for a URL |

```bash
reddit cache info
```

## Meta

| Command | Does |
|---|---|
| `open <id\|url>` | Open a Reddit page in the default browser |
| `info` | Show configuration, paths, and the affiliation disclaimer |
| `version` | Print version, commit, and build date |

```bash
reddit open 1abc23
```

## Global flags

These apply to every command. See [configuration](/reference/configuration/) for
the full list and their defaults.

| Flag | Meaning |
|---|---|
| `-o, --output` | Output format (default table on a TTY, jsonl piped) |
| `--fields` | Comma-separated fields to show |
| `--no-header` | Omit the header row in table/csv/tsv output |
| `--template` | Go text/template applied per record |
| `--color` | `auto`, `always`, or `never` |
| `-n, --limit` | Maximum rows (`0` means all) |
| `--pages` | Listing pages to walk (`0` means until exhausted or limit) |
| `-q, --quiet` | Suppress progress output |
| `-j, --workers` | Concurrency (default 2) |
| `--delay` | Minimum delay between requests (default 2s) |
| `--timeout` | Per-request timeout (default 30s) |
| `--retries` | Retry attempts on 429/5xx (default 3) |
| `--cache-ttl` | Cache lifetime (default 24h) |
| `--no-cache` | Bypass the on-disk cache for this run |
| `--refresh` | Force a re-fetch, ignoring the cache |
| `--data-dir` | Root data directory (env `REDDIT_DATA_DIR`) |
| `--store` | SQLite store path (default `<data-dir>/reddit.db`) |
| `--user-agent` | User-Agent sent with each request |
| `--cookies` | Netscape cookie jar to lend a signed-in session |
