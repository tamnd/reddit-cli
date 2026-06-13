---
title: "Output formats"
description: "Every output format, how to narrow fields, and how to template records."
weight: 30
---

Every command that emits records renders through the same formatter. Pick a
format with `--output` (or `-o`), or let reddit choose: a table when writing to a
terminal, JSONL when piped.

## Formats

```bash
reddit posts golang -o table   # aligned columns for reading
reddit posts golang -o jsonl   # one JSON object per line, for piping
reddit posts golang -o json    # a single JSON array
reddit posts golang -o csv     # spreadsheet friendly
reddit posts golang -o tsv     # tab-separated
reddit posts golang -o url     # just the URL of each row
reddit posts golang -o raw     # the underlying bytes, unformatted
```

| Format | Best for |
|---|---|
| `table` | Reading on a terminal |
| `jsonl` | Piping into another tool, one object at a time |
| `json` | Loading a whole result as an array |
| `csv` / `tsv` | Spreadsheets and quick column math |
| `url` | Feeding URLs into other commands |
| `raw` | The unformatted bytes |

## Narrowing fields

Keep only the fields you want:

```bash
reddit posts golang --fields title,score,num_comments
```

`--no-header` drops the header row in `table`, `csv`, and `tsv` output, which is
handy when a downstream tool expects bare rows.

## Templating records

For full control over each line, apply a Go text/template. The fields are the
record's keys:

```bash
reddit posts golang --template '{{.title}} ({{.score}})'
reddit comments 1abc23 --template '{{.author}}	{{.body}}'
```

## Why auto-detection helps

Because the default adapts to the destination, the same command reads well by
hand and parses cleanly in a pipe:

```bash
reddit posts golang                  # a table, because this is a terminal
reddit posts golang | jq -r .url     # JSONL, because this is a pipe
```

You only reach for `--output` when you want something other than that default.

## Color

`--color` is `auto` by default: reddit colors table output on a terminal and
drops color when piped. Force it with `--color always` or turn it off with
`--color never`.
