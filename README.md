# What is this?

This is a Linux/D-Bus notification daemon for MPD, written in Go.

# How to install?

This code uses [Go modules](https://github.com/golang/go/wiki/Modules),
so you'll need Go 1.11.

Clone this somewhere, `cd` into its directory, and:

```bash
go mod vendor
go build
```

## Unsupported: GOPATH

The old `go get` build method may work, but I'm probably going to rip
out the `GOPATH` junk out of my `zsh` dotfiles once I finish writing
this, if you catch my drift...  Installation this way was easier
(assuming you were OK with Go's opinions on where to put things):

```bash
go get github.com/jws85/mpdnotify
```

# Attributions

["Music Note" icon](https://thenounproject.com/search/?q=music%20note&i=1060111)
by [Björn Andersson](https://thenounproject.com/bjorna1) is licensed under
[CC BY 3.0](https://creativecommons.org/licenses/by/3.0/us/legalcode).
