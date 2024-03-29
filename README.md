# Sleeping Beauty

From Wikipedia:

> "Sleeping Beauty" (French: La Belle au centre de données dormant),
> or "Little Briar Rose" (German: Dornröschen), also titled in English
> as "The Sleeping Beauty in the Datacenter", is a classic fairy tale
> about a web application who is cursed to sleep for an indeterminate
> period of time by an evil systems administrator, to be awakened by a
> handsome TCP request at the end of it.

\[citation needed\]

## Status

Sleeping Beauty is considered largely feature-complete, stable, and
suitable for use in production. New features may be added in the case
of compelling reported use cases, but backwards compatibility will be
preserved except on major version releases, see the changelog for
details. Bug fixes and security issues will be addressed promptly to
the best of my ability.

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

# Optional. Port on which Sleeping Beauty will expose metrics. No
# default value; if not provided then a metrics server is not run. You
# can access pprof profiling data at /debug/pprof, and Prometheus
# metrics at /metrics.
SLEEPING_BEAUTY_METRICS_PORT=9090

# Optional. Network interface on which to expose metrics. Defaults to
# 0.0.0.0, meaning listen on all interfaces. You may wish to set this
# to 127.0.0.1 if your metrics are ingested by a sidecar process
# running in the container.
SLEEPING_BEAUTY_METRICS_HOST=0.0.0.0
```

After configuring environment variables, simply run the `sleepingd`
binary. It will listen on the specified port, and will not terminate
until sent a signal. You can verify operation by making a request to
the `SLEEPING_BEAUTY_LISTEN_PORT` on localhost with curl, and
observing the logs and HTTP response.

## Installation

Sleeping Beauty is distributed as a single, statically-linked binary.
To install it, download from GitHub Releases your preferred artifact
(tarball, deb, rpm, apk) for your preferred machine architecture,
operating system, and version, and install to taste. You can also pull
from [Docker
Hub](https://hub.docker.com/r/radiansoftware/sleeping-beauty) which
may be convenient for copying the binary into your Dockerfile.

Alternatively, you may compile your own binary. This is done by
cloning this repository at the desired revision, and running `make
build`. You have to have Go installed.

## Containerization

If running Sleeping Beauty in a containerized environment (e.g.
Docker) then it is your responsibility to supply an [appropriate
pid1](https://blog.phusion.nl/2015/01/20/docker-and-the-pid-1-zombie-reaping-problem/)
to ensure that zombie processes are reaped properly. Modern versions
of Docker can accomplish this transparently if you pass `--init` to
`docker run`.

## Caveats

There's one major problem with Sleeping Beauty, which is that there is
unfortunately no way for it to guarantee that all resources are
released when it shuts down your server process.

One common way that this can fail is if your server process starts its
own subprocesses, which are not properly terminated when the parent
dies. You can check the output of `ps` before and after Sleeping
Beauty terminates your server, to see that all new sub-processes have
terminated.

The other big issue has to do with the [kernel file
cache](https://unix.stackexchange.com/q/736941). When you read in
files from disk, the kernel keeps them in memory. As a general
practice, the kernel keeps lots of things in memory, often filling it
to its limit, because the vast majority of that memory can be easily
freed as soon as it is needed for something else (thus it would be a
waste of time to bother doing it earlier than necessary). However,
it's an issue because tools for container resource utilization (such
as `docker stats`) often take into account cached memory even when it
is freeable.

So, for example, if your webserver reads a bunch of files, then is
terminated by Sleeping Beauty, some memory may be retained even if the
memory-holding processes in the container are all terminated.
Depending on the workload, this overhead can be substantial. You can
check the hypothesis by investigating the `docker stats` output
before, during, and after the execution of your server process by
Sleeping Beauty. One way to reclaim some memory is by running `vmtouch
-e` on modified files and directories. You can typically reclaim *all*
cached memory with `echo 3 > /proc/sys/vm/drop_caches`, but this can
only be done as root from the host system, not from within a
container.

As a result, you can run into issues in environments like
[Railway](https://railway.app/) where you are billed based on measured
resource utilization. I reached out to them about discrepancies in
memory metrics leading to higher bills, and they confirmed that this
is a known problem, but it's hard to solve given the limits of
containerization tooling at present.

As always when deploying applications into production, exercise
caution and monitor your metrics.

## Run tests

Execute `make test-unit` (requires Go) or `make test-integration`
(requires Docker) or `make test` to do both.
