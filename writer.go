package slackhook

import (
	"bytes"
	"fmt"
	"github.com/azer/logger"
	"net/http"
	"time"
)

type Writer struct {
	WebHookURL   string
	Channel      string
	Username     string
	Filter       func(*logger.Log) bool
	Queue        []string
	LastPostedAt int64
}

func (writer *Writer) ClearQueue() string {
	content := ""
	for _, row := range writer.Queue {
		content = fmt.Sprintf("%s%s\n", content, row)
	}

	writer.Queue = []string{}

	return content
}

func (writer *Writer) Post(log *logger.Log) {
	writer.Queue = append(writer.Queue, fmt.Sprintf("%s %s %s", writer.FormatLevel(log.Level), log.Message, writer.FormatAttrs(log.Attrs)))

	// Post slack every with 10 seconds breaks
	if Now()-writer.LastPostedAt <= 10 {
		// Not ready yet, keep it in the queue.
		return
	}

	writer.LastPostedAt = Now()
	content := writer.ClearQueue()

	var body = []byte(fmt.Sprintf(`{"channel": "#%s", "username": "%s", "text": "%s"}`, writer.Channel, writer.Username, content))

	req, err := http.NewRequest("POST", writer.WebHookURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(fmt.Sprintf("logger-slack-hook error: %v", err))
	}

	defer resp.Body.Close()
}

func (writer Writer) Write(log *logger.Log) {
	if writer.Filter != nil {
		if writer.Filter(log) {
			writer.Post(log)
		}
	} else {
		writer.Post(log)
	}
}

func (writer *Writer) FormatAttrs(attrs *logger.Attrs) string {
	if attrs == nil {
		return ""
	}

	result := ""
	for key, val := range *attrs {
		result = fmt.Sprintf("%s\n       %s: %v", result, key, val)
	}

	return result
}

func (writer *Writer) FormatLevel(level string) string {
	switch level {
	case "INFO":
		return ":memo:"
	case "ERROR":
		return ":warning:"
	case "TIMER":
		return ":timer_clock:"
	default:
		return ":no_mouth:"
	}
}

func Now() int64 {
	return time.Now().Unix()
}
