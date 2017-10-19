package slackhook

import (
	"bytes"
	"fmt"
	"github.com/azer/logger"
	"net/http"
	"strings"
	"time"
)

const DEFAULT_INTERVAL_SECS = 10

type Writer struct {
	WebHookURL      string
	Channel         string
	Username        string
	Filter          func(*logger.Log) bool
	Queue           []string
	LastPostedAt    int64
	IsWorkerRunning bool
	IntervalSecs    int
}

func (writer *Writer) ClearQueue() string {
	content := ""
	for _, row := range writer.Queue {
		content = fmt.Sprintf("%s%s\n", content, row)
	}

	writer.Queue = []string{}

	return content
}

func (writer *Writer) Append(log *logger.Log) {
	writer.Queue = append(writer.Queue, fmt.Sprintf("%s %s %s", writer.FormatLevel(log), log.Message, writer.FormatAttrs(log.Attrs)))

	if !writer.IsWorkerRunning {
		go writer.Worker()
	}
}

func (writer *Writer) Worker() {
	if writer.IsWorkerRunning {
		return
	}

	writer.IsWorkerRunning = true
	if writer.IntervalSecs == 0 {
		writer.IntervalSecs = DEFAULT_INTERVAL_SECS
	}

	t := time.NewTicker(time.Duration(writer.IntervalSecs) * time.Second)
	for {
		writer.Post()
		<-t.C
	}

	writer.IsWorkerRunning = false
}

func (writer Writer) Write(log *logger.Log) {
	if writer.Filter != nil {
		if writer.Filter(log) {
			writer.Append(log)
		}
	} else {
		writer.Append(log)
	}
}

func (writer *Writer) FormatAttrs(attrs *logger.Attrs) string {
	if attrs == nil {
		return ""
	}

	result := ""
	for key, val := range *attrs {
		result = fmt.Sprintf("%s\n %s: %v", result, key, val)
	}

	if len(strings.TrimSpace(result)) > 0 {
		return fmt.Sprintf("```%s```", result)
	}

	return ""
}

func (writer *Writer) FormatLevel(log *logger.Log) string {
	switch log.Level {
	case "INFO":
		return ":memo:"
	case "ERROR":
		return ":mushroom:"
	case "TIMER":
		return fmt.Sprintf(":turtle: %dms :turtle:", log.Elapsed)
	default:
		return ":no_mouth:"
	}
}

func (writer *Writer) Post() {
	writer.LastPostedAt = Now()
	content := writer.ClearQueue()

	var body = []byte(fmt.Sprintf(`{"channel": "#%s", "username": "%s", "text": "%s"}`, writer.Channel, writer.Username, content))

	req, err := http.NewRequest("POST", writer.WebHookURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(fmt.Sprintf("logger-slack-hook error: %v", err))
	} else {
		defer resp.Body.Close()
	}
}

func Now() int64 {
	return time.Now().Unix()
}
