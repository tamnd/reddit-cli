---
title: "Community metadata"
description: "Pull a subreddit's rules, moderators, and wiki pages, and the other discussions of a link."
weight: 40
---

A community is more than its listing. reddit reads the metadata around it: the
posted rules, the moderator list, the wiki, and the duplicate discussions that
point at the same link.

## Rules

```bash
reddit rules golang
```

Each record is one posted rule: its short name, the full description, what it
applies to (posts, comments, or both), and its priority. This is the same list a
community shows in its sidebar.

## Moderators

```bash
reddit mods golang
```

Each record is one moderator: the name, the full name, when they were added, and
the permissions they hold. The list is public for any non-private community.

## Wiki

```bash
reddit wiki golang            # the wiki index page
reddit wiki golang faq        # a named page
```

`wiki` fetches a page's content (the markdown source), who last revised it, and
whether the current viewer may revise it. With one argument it reads the index;
a second argument names a specific page.

To see what pages a wiki has, list its index:

```bash
reddit wiki-pages golang
```

Each row is a page path you can hand back to `wiki`.

## Duplicates

When the same link gets posted to several communities, Reddit tracks the other
discussions. `duplicates` lists them:

```bash
reddit duplicates 1abc23 -n 20
```

The argument is a post id or URL, and each record is another post of the same
link, so you can see every community that discussed it. `dupes` is an alias.

## When metadata is gated

A private or quarantined community answers these with a block or a 403, and
reddit exits with code 5 (blocked) or 3 (no data) accordingly. A community with
no wiki, or one that restricts its wiki to moderators, returns nothing to read.
See [troubleshooting](/reference/troubleshooting/) for what each exit code means.
