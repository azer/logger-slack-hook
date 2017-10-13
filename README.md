# logger-slack-hook

Slack hook for [logger](https://github.com/azer/logger).
You stream specific packages, log levels and logs with specific attributes to Slack.
For example, you can log MySQL queries taking longer than 500ms into Slack:

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
