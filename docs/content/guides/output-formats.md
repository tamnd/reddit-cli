---
title: "Output formats"
description: "Render records as a table, JSON, CSV, or your own template, and script against the exit codes."
weight: 60
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

## Narrowing columns

Keep only the fields you want:

```bash
reddit posts golang --fields title,score,num_comments
reddit comments 1abc23 --fields author,depth,body
```

`--no-header` drops the header row in `table`, `csv`, and `tsv` output, which is
handy when a downstream tool expects bare rows.

## Templating rows

For full control over each line, apply a Go text/template. The fields are the
record's keys:

```bash
reddit posts golang --template '{{.title}} ({{.score}})'
reddit comments 1abc23 --template '{{.author}}	{{.body}}'
```

## Piping

Because the default adapts to the destination, the same command reads well by
hand and parses cleanly in a pipe:

```bash
reddit posts golang                  # a table, because this is a terminal
reddit posts golang | jq -r .url     # JSONL, because this is a pipe
```

`--limit` (or `-n`) caps the number of rows; `0` means all (up to one page,
unless `--pages` walks more).

## Exit codes for scripting

reddit returns a stable exit code so a script can branch on the outcome:

| Code | Meaning |
|---|---|
| `0` | OK |
| `1` | Error |
| `2` | Usage error |
| `3` | No data (nothing matched) |
| `4` | Partial (some items failed) |
| `5` | Blocked (rate-limited or a block page) |

For example, treat a block differently from a real failure:

```bash
reddit posts golang -o json > golang.json
case $? in
  0) echo "got it" ;;
  5) echo "blocked, slow down or pass --cookies" ;;
  *) echo "failed" ;;
esac
```

See [troubleshooting](/reference/troubleshooting/) for what to do about exit 5
and exit 3.
