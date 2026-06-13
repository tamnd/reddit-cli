---
title: "Introduction"
description: "What reddit reads, how the .json view turns a page into a record, and the rate-limit reality it stays inside."
weight: 10
---

[Reddit](https://www.reddit.com) is a large public site of communities, posts,
comment threads, and user profiles. It has an official API, but using it means
registering an app, holding a token, and living inside per-app quotas. There is
a simpler door for read-only work: every public Reddit page also serves a `.json`
view of the same content.

reddit is built on that door. It is a single binary that fetches the public JSON
view of a Reddit page and turns it into a structured record. You ask for a
listing, a comment tree, a user, or a subreddit, and it hands you fields, not
HTML and no token.

## The .json view

Append `.json` to almost any Reddit URL and you get the same content as
structured data: `/r/golang/top.json` is the top listing,
`/comments/1abc23.json` is a post and its comment tree,
`/user/spez/about.json` is a profile. reddit knows the shape of each of these
endpoints and the pagination they use, so you give it a name or a URL and it
walks the right one for you.

## Reddit's "thing" types

Everything on Reddit is a "thing" with a type prefix: `t1` is a comment, `t2` is
an account, `t3` is a post (a "link"), and `t5` is a subreddit. A full name like
`t3_1abc23` is the type plus the id. reddit uses these the same way Reddit does:
the `id` command classifies any URL or id into its (kind, id) pair without
touching the network, and a listing carries each child's kind so a mixed result
sorts itself out.

## From a listing to records

A Reddit listing is a page of children plus an `after` cursor for the next page.
reddit walks that cursor for as many pages as you ask for with `--pages` (or
until a `--limit` is met), trims each child to a clean record, and renders the
batch. A comment tree is nested rather than flat, so reddit flattens it into one
record per comment, each carrying its `depth` and `parent` so the shape survives
the flattening. The collapsed "load more" stubs in a deep thread are followed
only when you pass `--expand`.

## Polite by default, and the block reality

reddit waits two seconds between requests and runs two workers by default, so a
busy session stays a good citizen against a public site. Reddit rate-limits
aggressively when a client goes too fast or sends a generic agent, so reddit
sends a descriptive default User-Agent and backs off on a 429.

When Reddit decides a request is too much, it answers with a rate-limit page, a
`403`, or its "whoa there, pardner" interstitial instead of the content. reddit
recognises those and exits with code 5 ("blocked") rather than pretending it got
data. Slowing down with `--delay`, or lending a signed-in session with
`--cookies`, usually clears it. Datacenter and shared IPs are blocked the
hardest; a normal home or office connection rarely sees this at the default
rate.

## Independent and public-data only

reddit is an independent, open-source tool. It is not affiliated with, endorsed
by, or sponsored by Reddit, Inc. It reads only public pages, at a polite default
rate. It does not log in for you, store your credentials, or touch anything
behind an account.

Next: [install it](/getting-started/installation/), then take the
[quick start](/getting-started/quick-start/).
