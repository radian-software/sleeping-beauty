# Sleeping Beauty

From Wikipedia:

> "Sleeping Beauty" (French: La Belle au centre de données dormant),
> or "Little Briar Rose" (German: Dornröschen), also titled in English
> as "The Sleeping Beauty in the Datacenter", is a classic fairy tale
> about a web application who is cursed to sleep for an indeterminate
> period of time by an evil systems administrator, to be awakened by a
> handsome TCP request at the end of them.

\[citation needed\]

## Synopsis

Sleeping Beauty allows you to run a web application that automatically
idles when not in use, which can save resources for low-traffic
applications.

You provide the shell command that starts your web application
listening on localhost. Sleeping Beauty starts a small proxy server
that listens on another port you specify, which can be exposed
externally. When TCP traffic arrives, Sleeping Beauty starts your
application in the background, waits for it to become healthy, and
then proxies traffic transparently. After a configurable time of no
traffic through the proxy, Sleeping Beauty can automatically terminate
your application. (If another request comes in unexpectedly, it is
held until the application can be re-started automatically; in other
words, connection draining is handled gracefully.)

## Usage

Sleeping Beauty is distributed as a single, statically-linked binary
which can be configured using environment variables:

```bash
# Required. Passed to the default shell for the current user as a
# single command string to execute with '-c'. No default value.
SLEEPING_BEAUTY_COMMAND="node server.js"

# Required. Number of seconds to wait with no TCP traffic after which
# to shut down the application. No default value.
SLEEPING_BEAUTY_TIMEOUT_SECONDS=60

# Required. Port of the webserver that is launched by running the
# shell command you provided. This should be listening on localhost.
# No default value.
SLEEPING_BEAUTY_COMMAND_PORT=8080

# Required. Port on which Sleeping Beauty will listen for incoming
# connections. No default value. If this is a well-known port then
# Sleeping Beauty will need to be run as root.
SLEEPING_BEAUTY_LISTEN_PORT=80

# Optional. Network interface on which Sleeping Beauty will listen for
# incoming connections. Defaults to 0.0.0.0, meaning listen on all
# interfaces. You may wish to set this to 127.0.0.1 instead if you
# have placed Sleeping Beauty behind a further proxy or load balancer.
SLEEPING_BEAUTY_LISTEN_HOST=0.0.0.0

# Optional. Path to request on the local server to detect when it has
# finished starting up. Requests are made once per second,
# synchronously, until one returns 2xx. Pending connections will be
# held until the server passes health checks. If absent or empty, no
# healthcheck is done and pending requests are held for one second
# before being passed through.
SLEEPING_BEAUTY_HEALTHCHECK_PATH=/health
```

After configuring environment variables, simply run the `sleepingd`
binary. It will listen on the specified port, and will not terminate
until sent a signal. You can verify operation by making a request to
the `SLEEPING_BEAUTY_LISTEN_PORT` on localhost with curl, and
observing the logs and HTTP response.

## Installation

Sleeping Beauty is distributed as a single, statically-linked binary.
To install it, download the binary for your preferred version from
GitHub Releases, and install it to your preferred location.

Alternatively, you may compile your own binary. This is done by
cloning this repository at the desired revision, and running `go build
./cmd/sleepingd`.
