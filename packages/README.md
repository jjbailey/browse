## Packages for browse

This directory contains signed packages for browse. The packages contain a single static binary called `browse` and a symlink to it called `br`. The commands install in `/usr/local/bin`.

## Build

To build browse:

    $ cd /path/to/go/src
    $ go build -ldflags="-linkmode external -extldflags -static -s -w" -trimpath .
