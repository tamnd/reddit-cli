---
title: "Profiles and search"
description: "Look up users and subreddits, list a user's posts and comments, search posts, and discover communities and people."
weight: 30
---

Beyond a single listing, reddit reads profiles and runs the site's search. All
of it comes off the same public `.json` view.

## User and subreddit profiles

```bash
reddit user spez
reddit subreddit golang
```

Both take a bare name, a `u/` or `r/` prefix, or a full URL. A user record
carries the name and full name, the account age, link and comment karma, and the
gold and verified flags. A subreddit record carries the name and full name, the
title and public description, the subscriber count, the created date, the type
(public, restricted, private), and the over-18 flag.

Both take several names at once:

```bash
reddit subreddit golang rust zig -o csv
```

`subreddit` answers to `sub` and `r`; `user` answers to `u`.

## A user's posts and comments

```bash
reddit user-posts spez -n 50
reddit user-comments spez -n 50
```

`user-posts` lists what a user submitted; `user-comments` lists what they wrote.
Both page like any listing: cap with `-n`, walk pages with `--pages`.

## Searching posts

```bash
reddit search "context deadline"               # site-wide
reddit search "generics" --in golang -n 25      # within one community
```

`--in` restricts the search to a subreddit. Sort and window it:

```bash
reddit search "release" --sort new
reddit search "best of" --sort top --time year
```

`--sort` takes `relevance`, `hot`, `top`, `new`, or `comments`; `--time` takes
the usual windows. `--nsfw` includes over-18 results, which are filtered out by
default.

## Discovering communities and people

`subreddits` and `users` search the directories by name rather than searching
content:

```bash
reddit subreddits programming -n 20
reddit users gallowboob -n 20
```

`subreddits` returns community records (name, subscribers, description);
`users` returns account records.

## Classifying without fetching

`id` turns any URL or id into its (kind, id) pair offline, which is the safe way
to branch in a script before spending a request:

```bash
reddit id t3_1abc23
reddit id u/spez
reddit id https://www.reddit.com/r/golang/comments/1abc23/title/
```

```
link	1abc23
account	spez
link	1abc23
```

The kinds follow Reddit's "thing" types: `comment`, `account`, `link` (a post),
and `subreddit`. `open` resolves the same input to a URL and opens it in your
browser:

```bash
reddit open 1abc23
```
