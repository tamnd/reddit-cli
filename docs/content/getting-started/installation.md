---
title: "Installation"
description: "Install reddit from a release, with go install, from a package manager, or as a container."
weight: 20
---

## Prebuilt binaries

Every [release](https://github.com/tamnd/reddit-cli/releases) carries archives
for Linux, macOS, Windows, and FreeBSD on amd64 and arm64, plus deb, rpm, and
apk packages for Linux. Download, unpack, put `reddit` on your `PATH`, done. The
`checksums.txt` on each release is signed with keyless
[cosign](https://docs.sigstore.dev/), and SBOMs ship alongside, if you want to
verify before running.

## With Go

```bash
go install github.com/tamnd/reddit-cli/cmd/reddit@latest
```

That puts `reddit` in `$(go env GOPATH)/bin`, which is `~/go/bin` unless you
moved it. Make sure that directory is on your `PATH`.

## Homebrew

```bash
brew install --cask tamnd/tap/reddit
```

## Scoop

```bash
scoop bucket add tamnd https://github.com/tamnd/scoop-bucket
scoop install reddit
```

## Container image

The multi-arch image is on GHCR:

```bash
docker run --rm ghcr.io/tamnd/reddit:0.1.0 posts golang
```

## From source

```bash
git clone https://github.com/tamnd/reddit-cli
cd reddit-cli
make build        # produces ./bin/reddit
./bin/reddit version
```

## Requirements

- **Go 1.26 or later** to build. The released binary has no Go requirement.

That is the whole list. The binary is pure Go (CGO_ENABLED=0), so there is no
config file to write, no database to provision, and nothing to link against.

## Checking the install

```bash
reddit version
```

prints the version and exits. Then confirm it can reach Reddit:

```bash
reddit posts golang -n 3
```

should print a few posts from r/golang. If you see them, you are ready for the
[quick start](/getting-started/quick-start/).
