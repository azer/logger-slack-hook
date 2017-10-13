package main

import (
	"github.com/azer/logger"
	"github.com/azer/logger-slack-hook"
	"time"
)

func main() {

	// First setup a hook and tell the hook the rules of streaming to the hook.
	logger.Hook(&slackhook.Writer{
		// Create a webhook URL in Slack and enter it here.
		WebHookURL: "https://hooks.slack.com/services/...",
		Channel:    "slow-queries",
		Username:   "Query Person",
		// We don't want to log everything to Slack. It'd be too noisy. We need a filter.
		Filter: func(log *logger.Log) bool {
			// Log only MySQL package, only timers and only the ones taking longer than 500ms
			// So we'll get to know what queries needs to be optimized.
			return log.Package == "mysql" && log.Level == "TIMER" && log.Elapsed >= 500
		},
	})

	// Let's test the hook above.
	mysql := logger.New("mysql")
	images := logger.New("images")

	// Make some noooiseee
	timer := mysql.Timer()
	time.Sleep(time.Millisecond * 30)
	timer.End("Took 30 milliseconds to run this query... not bad")

	images.Info("I just created an image meanwhile... yay")

	timer = mysql.Timer()
	time.Sleep(time.Millisecond * 550)
	timer.End("A slow query, took 550ms. developers should be notified.", logger.Attrs{
		"foo": "bar",
		"qux": 123,
	})

	mysql.Error("Something produced an error, should not be logged though.")
	images.Error("Wow, errors everywhere.")

	timer = mysql.Timer()
	time.Sleep(time.Millisecond * 650)
	timer.End("Another slow query (650ms) what's happening here?")

	images.Info("Ok we're done here")
	mysql.Info("Good bye")
}
