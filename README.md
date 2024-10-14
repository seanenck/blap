b(inary) d(ownloader)
===

A simplistic "manager" of binaries downloaded from github (or other sources).
This is meant to be a simple, specific (and mostly for personal use) alternative
to local package management, go/cargo options, or others (e.g. eget).

[![build](https://github.com/seanenck/bd/actions/workflows/build.yml/badge.svg)](https://github.com/seanenck/bd/actions/workflows/build.yml)

# build

clone repository and:
```
make
```

to install
```
make install DESTDIR=~/chosen/path
```

# usage

Utilize the primitive `help` to see the CLI/config options for `bd`

## Config

Simple configuration for installing `go`, `rg`, and `nvim` locally
```
directory: ~/.local/fs
applications:
  go:
    mode: "tagged"
    remote:
      upstream: "https://github.com/golang/go"
      download: "https://go.dev/dl/{{ $.Tag }}.linux-amd64.tar.gz"
      filters: 
        - "refs/tags/weekly"
        - "refs/tags/release"
        - "[0-9]rc[0-9]"
    binaries:
      files:
        - "bin/go"
      destination: "~/.local/bin/developer"
  rg:
    mode: "github"
    remote:
      upstream: "BurntSushi/ripgrep"
      asset: "x86_64-unknown-linux-(.+?).tar.gz$"
    binaries:
      files:
        - "rg"
      destination: "~/.local/bin"
  nvim:
    mode: "github"
    remote:
      upstream: "neovim/neovim"
      asset: "nvim-linux64.tar.gz$"
    binaries:
      files:
        - "bin/nvim"
      destination: "~/.local/bin"
```
