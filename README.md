# subsonic-webui

A feature-incomplete subsonic music "player"

It... exists, unfortunately.

At this point, its more of a proxy, than a strict client

# Installation

1. Install [Go](https://go.dev/)
2. Clone this repo, and `cd` in.
3. Run `go build`, should spit out an executable binary named `subsonic-webui`, see usage

# Usage

This expects at minimum 3 parameters set from environment variables:

- SUBSONIC_SERVER: The scheme + domain for a given subsonic compatible server
- SUBSONIC_USERNAME: The username
- SUBSONIC_PASSWORD: The password

For example, in order:
- https://demo.navidrome.org
- demo
- demo

With these set, run the binary from above:

`./subsonic-webui [localhost:8080]`

It takes an optional argument to configure the bound socket interface and port. By default (unsupplied arg), this is set to localhost:8080.

After starting, open `http://localhost:8080` in your browser of choice.

