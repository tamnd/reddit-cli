---
title: "Bulk crawling"
description: "Seed post URLs from listings, queue them, crawl into a local SQLite store, and export the records."
weight: 50
---

For more than a page at a time, reddit has a small pipeline: seed post URLs from
subreddit listings, enqueue them, crawl the queue into a local SQLite store, and
export what you collected. Everything lands in one database file under your data
dir.

## 1. Seed from listings

`seed` walks one or more subreddit listings and emits the comment-page URL of
every post it finds:

```bash
reddit seed golang --sort top --time week
```

It takes the same `--sort` and `--time` as `posts`, and the same `-n` and
`--pages` to control how deep it walks. With no `--enqueue` it just prints the
URLs, so you can pipe them somewhere else:

```bash
reddit seed golang rust --sort top -o jsonl | jq -r .url
```

## 2. Enqueue

Add `--enqueue` to put the discovered URLs into the crawl queue instead of just
printing them:

```bash
reddit seed golang --sort top --time week --enqueue
```

## 3. Crawl the queue

`crawl` drains the queue: it fetches each post, caches the page, and with
`--parse` also stores the post and its comments in the records table. `--max`
caps how many to process (`0` drains the whole queue). It uses the global
`--workers` and `--delay`, so it stays polite:

```bash
reddit crawl --max 50 --parse
```

It reports how many it processed and failed. Exit code 3 means nothing was
processed (an empty queue); exit code 4 means some failed (see
[troubleshooting](/reference/troubleshooting/)).

## 4. Inspect and export the store

`db` works with the local SQLite store:

```bash
reddit db info                          # summarize records and the queue
reddit db count post                    # how many posts are stored
reddit db get post 1abc23               # one stored record as JSON
reddit db export --type post --out posts.jsonl
reddit db vacuum                        # reclaim space
```

`db export` writes every stored record (or just one `--type`) to a file with
`--out`, or to stdout.

## The page cache

Every fetch goes through an on-disk cache (content-addressed, gzip), so a
re-crawl does not re-fetch pages that have not changed. `cache` manages it:

```bash
reddit cache info                                              # location, file count, size
reddit cache path https://www.reddit.com/comments/1abc23.json   # the cache file for a URL
reddit cache clear                                             # remove every cached page
```

The cache TTL defaults to 24 hours. Bypass it for one run with `--no-cache`, or
force a re-fetch with `--refresh`.

## The whole pipeline

Put together, collecting a slice of a community looks like this:

```bash
reddit seed golang --sort top --time month -n 200 --enqueue
reddit crawl --parse
reddit db export --type post --out golang.jsonl
```

The store and cache both live under the data dir; point that elsewhere with
`--data-dir` or `REDDIT_DATA_DIR`, and point the database file alone with
`--store`. See [configuration](/reference/configuration/).
