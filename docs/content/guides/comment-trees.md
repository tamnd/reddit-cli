---
title: "Comment trees"
description: "Flatten a post's discussion into records, sort it, cap its depth, and follow the collapsed load-more stubs."
weight: 20
---

A Reddit comment thread is a tree: comments hold replies, replies hold more
replies. `comments` reads that tree and flattens it into one record per comment,
keeping each comment's depth and parent so the shape survives the flattening.

## Reading a thread

```bash
reddit comments 1abc23
```

The argument is a post id or a full comment-page URL. Each record carries the
comment id and full name, the link and parent ids, the subreddit and author, the
body, the score, the created and edited timestamps, and the depth. A
depth-aware tool can rebuild the tree from `parent_id`; a flat reader can sort by
depth and read top to bottom.

## Sorting and capping

Comments sort the way the site offers:

```bash
reddit comments 1abc23 --sort top
reddit comments 1abc23 --sort new
reddit comments 1abc23 --sort qa
```

`--sort` takes `confidence` (the site default, "best"), `top`, `new`,
`controversial`, `old`, or `qa`. Cap how many you pull with `-n`:

```bash
reddit comments 1abc23 --sort top -n 200
```

`--depth` asks Reddit to return only the first N levels of the tree, which keeps
a huge thread manageable when you only want the top of the conversation:

```bash
reddit comments 1abc23 --depth 2
```

## Following "load more"

A deep or wide thread does not ship every comment in the first response. Reddit
collapses the tail into "load more" stubs that name the hidden comment ids.
`--expand` follows those stubs through the `morechildren` endpoint and folds the
extra comments into the result:

```bash
reddit comments 1abc23 --expand
```

This costs extra requests (one per batch of hidden ids), so it is off by default.
Leave it off for a quick read of the visible thread; turn it on when you need the
whole discussion.

## Piping a thread

Because each comment is its own record, a thread pipes cleanly. Pull every body
out of a thread:

```bash
reddit comments 1abc23 --sort top -o jsonl | jq -r .body
```

Count comments by author:

```bash
reddit comments 1abc23 -o jsonl | jq -r .author | sort | uniq -c | sort -rn
```
