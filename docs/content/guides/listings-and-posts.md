---
title: "Listings and posts"
description: "Walk a subreddit by any sort, page through it, and fetch individual posts by id or URL."
weight: 10
---

A subreddit listing is the bread and butter of Reddit, and `posts` reads it the
way the site does: pick a sort, pick a time window, and page through as deep as
you want.

## Listing a subreddit

```bash
reddit posts golang
```

That is r/golang's hot listing. The argument is forgiving: `golang`, `r/golang`,
`/r/golang`, or a full subreddit URL all work. Sort it like the site:

```bash
reddit posts golang --sort new
reddit posts golang --sort top --time week
reddit posts golang --sort controversial --time all
```

`--sort` takes `hot`, `new`, `top`, `rising`, or `controversial`. `--time`
applies to `top` and `controversial` and takes `hour`, `day`, `week`, `month`,
`year`, or `all`.

## More than one subreddit

`posts` takes several subreddits at once and concatenates the results:

```bash
reddit posts golang rust zig -o json
```

When one of several fails (a private or banned community, say), reddit notes it
on stderr and carries on with the rest, so one bad name does not sink the batch.

## Paging deeper

A listing page holds up to 100 posts. Two flags go past that:

- `-n, --limit` caps the total number of posts. Ask for more than one page worth
  and reddit walks the `after` cursor for you.
- `--pages` sets how many pages to walk explicitly (`0` means until the limit is
  met or the listing runs out).

```bash
reddit posts golang --sort top --time year -n 500     # walks ~5 pages
reddit posts golang --pages 3                          # exactly 3 pages
```

reddit stays polite between pages: the global `--delay` applies to each request.

## Fetching one post

`post` fetches the link itself (without its comments) by id or URL:

```bash
reddit post 1abc23
reddit post https://www.reddit.com/r/golang/comments/1abc23/title/
```

It takes several ids or URLs at once, same as `posts`. For the discussion under
a post, use [`comments`](/guides/comment-trees/).

## Useful fields

A post record carries the id and full name, the subreddit, the title and author,
the score and upvote ratio, the comment count, the created and edited
timestamps, the URL and permalink, the domain, the flair, and the
self/over-18 flags. It also carries the media URL for a video post, the source
URL of the first preview image, the gallery image URLs, and the
archived/pinned/original-content flags. Narrow to what you need with `--fields`:

```bash
reddit posts golang --sort top --fields title,score,num_comments,permalink
```

See [output formats](/reference/output/) for the full set of renderers.
