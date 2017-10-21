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

func (writer *Writer) Init() {
	writer.IsWorkerRunning = true
	go writer.Worker()
}

func (writer *Writer) ClearQueue() []string {
	rows := writer.Queue
	writer.Queue = []string{}
	return rows
}

func (writer *Writer) Append(log *logger.Log) {
	writer.AppendString(fmt.Sprintf("%s %s %s", writer.FormatLevel(log), log.Message, writer.FormatAttrs(log.Attrs)))
}

func (writer *Writer) AppendString(logs ...string) {
	for _, log := range logs {
		(*writer).Queue = append((*writer).Queue, log)
	}
}

func (writer *Writer) Worker() {
	if writer.IntervalSecs == 0 {
		writer.IntervalSecs = DEFAULT_INTERVAL_SECS
	}

	t := time.NewTicker(time.Duration(writer.IntervalSecs) * time.Second)
	for {
		if len(writer.Queue) > 0 {
			rows := writer.ClearQueue()
			if err := writer.Post(rows); err != nil {
				writer.AppendString(rows...)
			}
		}

		<-t.C
	}

	writer.IsWorkerRunning = false
}

func (writer *Writer) Write(log *logger.Log) {
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

func (writer *Writer) Post(rows []string) error {
	if len(rows) == 0 {
		return nil
	}

	writer.LastPostedAt = Now()
	content := StringifyRows(rows)

	var body = []byte(fmt.Sprintf(`{"channel": "#%s", "username": "%s", "text": "%s"}`, writer.Channel, writer.Username, content))

	req, err := http.NewRequest("POST", writer.WebHookURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	} else {
		defer resp.Body.Close()
	}

	return nil
}

func Now() int64 {
	return time.Now().Unix()
}

func StringifyRows(rows []string) string {
	content := ""
	for _, row := range rows {
		content = fmt.Sprintf("%s%s\n", content, row)
	}

	return content
}
