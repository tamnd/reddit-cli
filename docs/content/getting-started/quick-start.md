---
title: "Quick start"
description: "From an empty terminal to a real listing, a real comment tree, and a real profile, in a handful of commands."
weight: 30
---

This walks the core loop: list a subreddit, read a post's comment tree, look up
a user and a community, and search. Every command here reads the public `.json`
view, so none of it needs an API key or an account.

## 1. List a subreddit

```bash
reddit posts golang
```

Each row is a post from r/golang's hot listing. Sort and window it like the site
does:

```bash
reddit posts golang --sort top --time week -n 25
```

`--sort` takes `hot`, `new`, `top`, `rising`, or `controversial`; `--time`
applies to `top` and `controversial`.

## 2. Read a comment tree

```bash
reddit comments 1abc23
```

That flattens the discussion under post `1abc23` into one record per comment,
each carrying its depth and parent so the tree shape survives. Sort and cap it:

```bash
reddit comments 1abc23 --sort top -n 200
```

Deep threads hide their tail behind "load more" stubs. Follow them with
`--expand`:

```bash
reddit comments 1abc23 --expand
```

## 3. Look up a user and a subreddit

```bash
reddit user spez
reddit subreddit golang
```

Both take a bare name, a `u/` or `r/` prefix, or a full URL, and return a
structured record. The same as JSON:

```bash
reddit subreddit golang -o json
```

## 4. Search

```bash
reddit search "context deadline" --in golang -n 25
```

Drop `--in` to search site-wide. Discover communities and users by name:

```bash
reddit subreddits programming -n 20
reddit users gallowboob -n 20
```

## 5. Classify without fetching

`id` turns a URL or id into a (kind, id) pair without touching the network,
which is handy in scripts:

```bash
reddit id https://www.reddit.com/r/golang/comments/1abc23/title/
```

```
link	1abc23
```

## 6. Compose

Output that pipes is the point. Pull the URLs off a listing:

```bash
reddit posts golang --sort top -o jsonl | jq -r .url
```

Get one author's posts, keeping two fields:

```bash
reddit user-posts spez --fields title,score
```

## Where to next

You have the core loop. From here:

- [Listings and posts](/guides/listings-and-posts/) covers `posts`, `post`, and
  the sort and time windows.
- [Comment trees](/guides/comment-trees/) goes deep on `comments`, depth, and
  `--expand`.
- [Profiles and search](/guides/profiles-and-search/) covers `user`,
  `subreddit`, `search`, `subreddits`, `users`, and the `id` classifier.
- [Community metadata](/guides/community-metadata/) covers `rules`, `mods`,
  `wiki`, `wiki-pages`, and `duplicates`.
- The [CLI reference](/reference/cli/) lists every command and flag.
