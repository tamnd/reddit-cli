---
title: "reddit"
description: "A fast, friendly command line for public Reddit data. List a subreddit, read a comment tree, look up a user, search, and build datasets, all from one binary."
heroTitle: "Reddit, from the command line"
heroLead: "reddit is a single pure-Go binary that turns the public .json view of Reddit into clean, structured records. List a subreddit, read a comment tree, look up users and communities, search, and crawl in bulk, with no API key and nothing to sign up for."
heroPrimaryURL: "/getting-started/quick-start/"
heroPrimaryText: "Get started"
---

Every Reddit page has a `.json` twin: add `.json` to a URL and the same content
comes back as structured data instead of HTML. reddit is built on that. It reads
the public JSON view, classifies what it found through Reddit's "thing" types
(posts, comments, accounts, subreddits), and hands you records with real fields.

```bash
reddit posts golang --sort top --time week   # a subreddit listing
reddit comments 1abc23                        # a flattened comment tree
reddit user spez                              # a user profile as a record
reddit search "context deadline" --in golang  # search within a subreddit
```

It talks to `www.reddit.com` over plain HTTPS with no API key and no account.
The binary is pure Go with no runtime dependencies. Output is a table on a
terminal and JSONL when piped, with json, csv, tsv, url, raw, `--fields`, and
`--template` when you want something else.

## What you can do with it

- **Read listings.** `reddit posts` walks a subreddit by `hot`, `new`, `top`,
  `rising`, or `controversial`, across as many pages as you ask for.
- **Read comment trees.** `reddit comments` flattens a post's discussion into one
  record per comment, keeping depth and parent links, with `--expand` to follow
  the collapsed "load more" stubs.
- **Look up records.** Fetch a post, a user, or a subreddit by id or URL and get
  a structured record with scores, timestamps, flair, and cross-references.
- **Search and discover.** Search posts site-wide or inside one community, and
  discover subreddits and users by name.
- **Read community metadata.** Pull a subreddit's rules, moderators, and wiki
  pages, and the duplicate discussions of any link.
- **Crawl in bulk.** Seed post URLs from listings, queue them, crawl them into a
  local SQLite store, and export the records as JSONL.

## Independent and public-data only

reddit is an independent, open-source tool. It is not affiliated with, endorsed
by, or sponsored by Reddit, Inc. It reads only public pages, at a polite default
rate (a two second gap between requests, two workers).

## Where to go next

- New here? Start with the [introduction](/getting-started/introduction/) for
  the mental model, then the [quick start](/getting-started/quick-start/).
- Want to install it? See [installation](/getting-started/installation/).
- Looking for a specific task? The [guides](/guides/) cover listings, comment
  trees, profiles and search, community metadata, bulk crawling, and output.
- Need every flag? The [CLI reference](/reference/cli/) is the full surface.
