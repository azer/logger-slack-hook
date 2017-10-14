# logger-slack-hook

Slack hook for [logger](https://github.com/azer/logger).
Packages, log levels and attributes can be specified for streaming into Slack.
For example, you can get MySQL queries taking longer than 500ms reported to Slack:

```go
import (
  "github.com/azer/logger"
  "github.com/azer/logger-slack-hook"
)

func main () {
  logger.Hook(&SlackHook{
    WebHookURL: "https://hooks.slack.com/services/...",
    Channel: "slow-queries",
    Username: "Query Person",
    Filter: func (log *logger.Log) bool {
      return log.Package == "mysql" && log.Level == "TIMER" && log.Elapsed >= 500
    },
  })
}
```

See `examples/slow-queries.go` for working example.
