---
title: "Troubleshooting"
description: "The handful of things that trip people up, and how to fix each one."
weight: 40
---

Most of these come down to network reality, not a bug. Reddit is a public site
that rate-limits read traffic, and reddit is honest about what it can and cannot
read at a given moment.

## "blocked" and exit code 5

Reddit answers a request it judges too aggressive with a rate-limit page, a
`403`, or its "whoa there, pardner" interstitial instead of the content. When
reddit gets one of those, it exits with code 5 ("blocked") rather than returning
the block page as if it were data.

What to do, in order:

1. **Slow down.** The default `--delay` is two seconds and `--workers` is two. If
   you raised either, put them back. A blocked IP usually recovers after a short
   pause.
2. **Send a descriptive User-Agent.** reddit already sends one, but if you
   overrode it with a generic browser string, that is the most likely cause.
   Reddit rate-limits generic agents the hardest.
3. **Lend a session with `--cookies`.** Export a Netscape `cookies.txt` jar from
   a signed-in browser and pass it:

   ```bash
   reddit comments 1abc23 --cookies ~/cookies.txt
   ```

   A real session clears most blocks.

Datacenter, VPN, and shared IPs are blocked the hardest, often on the first
request, because Reddit treats them as bot traffic by default. A normal home or
office connection at the default rate rarely sees this. If every request from
your network blocks immediately regardless of delay, the IP itself is the cause,
and `--cookies` is the way through.

## The cookies.txt format

`--cookies` expects a Netscape cookie jar: the plain-text format most browser
extensions export and `curl` reads. Each line is tab-separated:

```
.reddit.com	TRUE	/	TRUE	0	reddit_session	abc123...
```

Lines starting with `#` are comments. Export it from a browser where you are
signed in to Reddit, save it somewhere private, and pass its path to
`--cookies`. reddit only replays the jar; it never logs in for you and never
stores credentials.

## "no data" and exit code 3

Exit code 3 means reddit reached the endpoint but found nothing to return: a
deleted post, an empty listing, a search with no matches, a wiki page that does
not exist. Check the id or URL is right (use `reddit id <url>` to see how reddit
classifies it), try a broader search, or confirm the thing you asked for still
exists.

## Rate limiting (429)

If Reddit returns 429 (too many requests), reddit backs off and retries up to
`--retries` times. If you see this often, you are going too fast: raise
`--delay`, lower `--workers`, and let the cache absorb repeat fetches. The
defaults (two second delay, two workers) are set to avoid this.

## Private, quarantined, and banned communities

A private subreddit answers with a `403` and reddit reports it as blocked. A
banned subreddit answers with a `404` and reddit reports no data. A quarantined
community may need a signed-in session that has opted in, which `--cookies`
provides. None of these is a tool fault; they are the access rules of the
community.

## A crawl reports failures (exit code 4)

`crawl` exits 4 when it processed some URLs but others failed (often a block on a
post in the queue). The records that did parse are in the store; re-run `crawl`
later to retry the queue, or pass `--cookies` for the blocked ones. Exit 3 from
`crawl` means nothing was processed at all (an empty queue).

## Where state lives

The on-disk cache and the SQLite store both live under the data dir (the XDG
data directory by default, or `REDDIT_DATA_DIR` / `--data-dir`). The store file
alone can be moved with `--store`. To see the resolved paths:

```bash
reddit info
```

To clear the cache and start fresh:

```bash
reddit cache clear
```
