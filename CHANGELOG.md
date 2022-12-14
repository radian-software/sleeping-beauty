# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog].

[keep a changelog]: https://keepachangelog.com/en/1.0.0/

## 4.1.0

Features:

* If you set `SLEEPING_BEAUTY_METRICS_PORT` to a positive integer,
  then Sleeping Beauty will expose pprof profiling data (at
  `/debug/pprof`) and Prometheus metrics (at `/metrics`) on the named
  port (bound to the interface named by
  `SLEEPING_BEAUTY_METRICS_HOST`, defaults to `0.0.0.0`).

Bugfixes:

* A memory leak has been fixed that would cause memory usage to
  increase over time when TCP connections were opened and closed to
  the server without any data being sent. This is believed to be a
  regression introduced in 4.0.0.

## 4.0.0

Behavior changes:

* The webserver is no longer awoken when a TCP connection is opened to
  it, only when some data is sent on that connection. This is
  technically a breaking change in behavior because there may be
  servers that want to send some data to a client as soon as they
  connect, but this is a better default for most cases since it
  doesn't actually make sense for queries like `nc -z` to wake the
  server (they are used by hosting providers like Railway to see if
  you have bound to the expected port, or as a health check).

Improvements:

* If something is already listening on the configured command port
  when Sleeping Beauty starts up, it will now immediately abort with
  nonzero exit status. This is surely desired, since if there is
  something already listening, then your webserver will not be able to
  start anyway when traffic comes in, and it will all get very
  confusing.

## 3.0.0

Behavior changes:

* The entire process group is signaled when turning off the webserver,
  not just the top-level process. This is helpful if your command is
  using something like bash which swallows signals. However it might
  be a breaking change so you should test that the new behavior works
  with your application before upgrading.

## 2.0.2

Bugfixes:

* We now correctly report upstream connection closure to clients, so
  that they know they need to open a new connection in that case even
  if they support long-term connection reuse. In practice this makes
  it so that web browsers such as Chrome and Firefox will not
  experience hangs when making new requests when working with a
  backend server that uses `Connection: keep-alive`.

## 2.0.1

Bugfixes:

* No more spurious errors like `read tcp
  127.0.0.1:39068->127.0.0.1:5001: use of closed network connection`
  logged when upstream server finishes sending data before client
  closes its connection (e.g., due to use of `Connection:
  keep-alive`).

## 2.0.0

Significant improvements:

* Webserver stays alive as long as there is active TCP traffic, even
  if there are not new connections being opened. This means Sleeping
  Beauty works better with long-lasting websockets or HTTP/2 pipes.
* Various race conditions fixed that could cause deadlocks, dropped
  connections, or fatal errors.
* Do not send back an error message when failing to connect to
  upstream endpoint. While a neat idea to report errors, it does not
  work well with most protocols built on top of TCP and will cause
  confusing downstream errors. Instead, the connection is closed
  immediately without a response when the upstream is not available.
  This behavior may be improved in future.
* Errors when proxying traffic are reported in logs. Improved log
  formatting in general.

## 1.0.0

Initial release. Configuration options:

* `SLEEPING_BEAUTY_COMMAND`
* `SLEEPING_BEAUTY_TIMEOUT_SECONDS`
* `SLEEPING_BEAUTY_COMMAND_PORT`
* `SLEEPING_BEAUTY_LISTEN_PORT`
* `SLEEPING_BEAUTY_LISTEN_HOST`
