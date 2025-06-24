package mper

import (
	"bytes"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	logLevel = parseLogLevel(os.Getenv("MP_LOG_LEVEL"))
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func parseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

func logf(level LogLevel, format string, v ...interface{}) {
	if level < logLevel {
		return
	}
	prefix := ""
	switch level {
	case DEBUG:
		prefix = "[DEBUG] "
	case INFO:
		prefix = "[INFO] "
	case WARN:
		prefix = "[WARN] "
	case ERROR:
		prefix = "[ERROR] "
	}
	log.Printf(prefix+format, v...)
}

type PullPushConfig struct {
	MetricsUrl          string `validate:"required" arg:"env:MP_METRICS_URL" help:"metrics url" default:""`
	PushgatewayUrl      string `validate:"required" arg:"env:MP_PG_URL" help:"pushgateway url" default:""`
	PushgatewayUsername string `arg:"env:MP_PG_USERNAME" help:"pushgateway username" default:""`
	PushgatewayPassword string `arg:"env:MP_PG_PASSWORD" help:"pushgateway password" default:""`
	PushgatewayCrontab  string `arg:"env:MP_PG_CRONTAB" help:"pushgateway crontab, default every 15 seconds" default:"*/15 * * * * *"`
}

func PullPushCrontab(config PullPushConfig) {
	log.Printf("[INFO] MP_LOG_LEVEL=%s, MP_METRICS_URL=%s, MP_PG_URL=%s, MP_PG_USERNAME=%s, MP_PG_PASSWORD=%s, MP_PG_CRONTAB=%s", os.Getenv("MP_LOG_LEVEL"), config.MetricsUrl, config.PushgatewayUrl, config.PushgatewayUsername, config.PushgatewayPassword, config.PushgatewayCrontab)
	if config.PushgatewayUrl != "" && config.MetricsUrl != "" {
		c := cron.New(cron.WithSeconds(), cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)), cron.WithLocation(time.UTC))

		logf(INFO, "crontab spec: %s", config.PushgatewayCrontab)
		c.AddFunc(config.PushgatewayCrontab, func() {
			PullPush(config)
		})
		c.Start()
	}
}

func PullPush(config PullPushConfig) {
	logf(INFO, "Prepare pull %s push to %s", config.MetricsUrl, config.PushgatewayUrl)

	resp, err := http.Get(config.MetricsUrl)
	if err != nil {
		logf(ERROR, "Error fetching metrics: %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		logf(ERROR, "Error fetching metrics: %d", resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logf(ERROR, "Error reading response body: %v", err)
		return
	}

	Push(config, body)
}

func Push(config PullPushConfig, data []byte) {
	logf(INFO, "Push to %s, username: %s", config.PushgatewayUrl, config.PushgatewayUsername)
	req, err := http.NewRequest("POST", config.PushgatewayUrl, bytes.NewBuffer(data))
	if err != nil {
		logf(ERROR, "Push error creating request: %v", err)
		return
	}

	if (config.PushgatewayUsername != "") && (config.PushgatewayPassword != "") {
		req.SetBasicAuth(config.PushgatewayUsername, config.PushgatewayPassword)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logf(ERROR, "Push error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	logf(INFO, "Push response status: %s", resp.Status)
}
