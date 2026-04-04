# core-go

> [!CAUTION]
> **This package is deprecated.** It is no longer supported and will not receive
> further updates. Please migrate to the new [cmd-stream-go v0.5.1](https://github.com/cmd-stream/cmd-stream-go) package.

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/core-go.svg)](https://pkg.go.dev/github.com/cmd-stream/core-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/core-go)](https://goreportcard.com/report/github.com/cmd-stream/core-go)
[![codecov](https://codecov.io/gh/cmd-stream/core-go/graph/badge.svg?token=RXPJ6ZIPK7)](https://codecov.io/gh/cmd-stream/core-go)

**core-go** contains the definitions for both `cmd-stream` client and server.

The client delegates all communication tasks, such as sending Commands,
receiving Results, and closing the connection to the `client.Delegate`.

The server manages client connections through `server.Delegate`, using a
configurable number of worker.
