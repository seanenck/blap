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

Simple configuration for installing `go`, `rg`, and `nvim` locally example is
[here](config.yaml)
