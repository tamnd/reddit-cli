---
title: "Configuration"
description: "The data directory, the store, cookies, the User-Agent, politeness knobs, environment, and exit codes."
weight: 20
---

reddit needs almost no configuration. There is no config file; every option is a
flag or an environment variable, and the defaults are chosen so the common case
needs neither. See everything reddit resolved with:

```bash
reddit info
```

It prints the configuration, the paths, and the affiliation disclaimer.

## The data directory

reddit keeps its state under one tree: the on-disk page cache and the SQLite
store. It defaults to the XDG data directory (for example
`~/.local/share/reddit` on Linux). Point it elsewhere with `--data-dir` or the
`REDDIT_DATA_DIR` environment variable.

## The store

The crawl pipeline writes records and the queue into a SQLite file, by default
`<data-dir>/reddit.db`. Point that single file somewhere else with `--store`,
which is handy when you want one corpus per project:

```bash
reddit crawl --parse --store ~/projects/golang/reddit.db
```

`db info`, `db count`, `db get`, `db export`, and `db vacuum` all read this file.

## The User-Agent

Reddit asks clients to send a descriptive, unique User-Agent and rate-limits
generic browser strings and empty agents the hardest. reddit sends
`reddit-cli (+https://github.com/tamnd/reddit-cli)` by default. Override it with
`--user-agent` if you are identifying your own bot:

```bash
reddit posts golang --user-agent "my-tool/1.0 (by /u/me)"
```

## Cookies

The `--cookies` flag takes a Netscape `cookies.txt` jar exported from a
signed-in browser session. reddit sends those cookies with each request, which
lends it a real session and usually gets past a block or rate-limit page. reddit
never logs in for you and never stores credentials; it only replays the jar you
hand it.

```bash
reddit comments 1abc23 --cookies ~/cookies.txt
```

See [troubleshooting](/reference/troubleshooting/) for the cookie file format.

## Caching

Every fetch goes through a content-addressed gzip cache on disk so a repeat run
does not re-fetch unchanged pages. `--cache-ttl` sets how long an entry stays
fresh (default 24h). `--no-cache` bypasses it for one run, and `--refresh`
forces a re-fetch and rewrites the entry. Manage the cache with `cache info`,
`cache path <url>`, and `cache clear`.

## Politeness

reddit is gentle by default so a busy session stays a good citizen against a
public site:

| Flag | Default | Meaning |
|---|---|---|
| `-j, --workers` | `2` | Concurrent requests |
| `--delay` | `2s` | Minimum gap between requests |
| `--timeout` | `30s` | Per-request timeout |
| `--retries` | `3` | Retry attempts on 429 and 5xx |

Raise `--workers` and lower `--delay` only when you have a reason to, and keep
them modest. Reddit rate-limits hard, and a slower run that finishes beats a
fast one that gets blocked.

## Environment variables

| Variable | Used for |
|---|---|
| `REDDIT_DATA_DIR` | Root data directory (overrides the XDG default) |

## Global flags

| Flag | Default | Meaning |
|---|---|---|
| `-o, --output` | auto | `table`, `json`, `jsonl`, `csv`, `tsv`, `url`, `raw` |
| `--fields` | all | Comma-separated fields to show |
| `--no-header` | off | Omit the header row in table/csv/tsv output |
| `--template` | none | Go text/template applied per record |
| `--color` | auto | `auto`, `always`, or `never` |
| `-n, --limit` | `0` | Maximum rows; `0` is all |
| `--pages` | `1` | Listing pages to walk; `0` is until exhausted or limit |
| `-q, --quiet` | off | Suppress progress output |
| `-j, --workers` | `2` | Concurrent requests |
| `--delay` | `2s` | Minimum delay between requests |
| `--timeout` | `30s` | Per-request timeout |
| `--retries` | `3` | Retry attempts |
| `--cache-ttl` | `24h` | Cache lifetime |
| `--no-cache` | off | Bypass the on-disk cache |
| `--refresh` | off | Force a re-fetch, ignoring the cache |
| `--data-dir` | XDG | Root data directory (env `REDDIT_DATA_DIR`) |
| `--store` | `<data-dir>/reddit.db` | SQLite store path |
| `--user-agent` | reddit-cli string | User-Agent sent with each request |
| `--cookies` | none | Netscape cookie jar |

## Output auto-detection

The default output format adapts to where it is going: an aligned table when the
output is a terminal, JSONL when it is piped. That keeps interactive use readable
and scripted use parseable without you setting `--output` either time. See
[output formats](/reference/output/) for the full set.

## Exit codes

reddit returns a stable exit code so scripts can branch on the outcome:

| Code | Meaning |
|---|---|
| `0` | OK |
| `1` | Error |
| `2` | Usage error |
| `3` | No data (nothing matched) |
| `4` | Partial (some items failed) |
| `5` | Blocked (rate-limited or a block page) |
