# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog].

[keep a changelog]: https://keepachangelog.com/en/1.0.0/

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
