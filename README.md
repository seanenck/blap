blap
===

A simplistic "manager" of binaries and source downloaded from github (or other sources)
for local deployment of tooling. This is meant to be a simple, specific (and mostly for personal use) alternative
to local package management (e.g. via go/cargo directly or tools like
eget/brew).

[![build](https://github.com/seanenck/blap/actions/workflows/build.yml/badge.svg)](https://github.com/seanenck/blap/actions/workflows/build.yml)

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

Utilize the primitive `help` to see the CLI/config options for `blap`

## Config

Simple configuration for installing `go`, `rg`, and `nvim` locally example is
[here](internal/config/examples)
