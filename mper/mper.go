package mper

import (
	"bytes"
	"fmt"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type PullPushConfig struct {
	MetricsUrl          string `validate:"required" arg:"env:METRICS_URL" help:"metrics url" default:""`
	PushgatewayUrl      string `validate:"required" arg:"env:PG_URL" help:"pushgateway url" default:""`
	PushgatewayUsername string `arg:"env:PG_USERNAME" help:"pushgateway username" default:""`
	PushgatewayPassword string `arg:"env:PG_PASSWORD" help:"pushgateway password" default:""`
	PushgatewayCrontab  string `arg:"env:PG_CRONTAB" help:"pushgateway crontab, default every 15 seconds" default:"*/15 * * * * *"`
}

func PullPushCrontab(config PullPushConfig) {
	if config.PushgatewayUrl != "" && config.MetricsUrl != "" {
		c := cron.New(cron.WithSeconds(), cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)), cron.WithLogger(
			cron.VerbosePrintfLogger(log.New(os.Stdout, "crontab: ", log.LstdFlags))), cron.WithLocation(time.UTC))

		log.Printf("crontab spec: %s", config.PushgatewayCrontab)
		c.AddFunc(config.PushgatewayCrontab, func() {
			PullPush(config)
		})
		c.Start()
	}
}

func PullPush(config PullPushConfig) {
	log.Printf("Prepare pull %s push to %s", config.MetricsUrl, config.PushgatewayUrl)

	resp, err := http.Get(config.MetricsUrl)
	if err != nil {
		fmt.Println("Error fetching metrics:", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error fetching metrics:", resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	Push(config, body)
}

func Push(config PullPushConfig, data []byte) {
	log.Printf("Push to %s, username: %s", config.PushgatewayUrl, config.PushgatewayUsername)
	req, err := http.NewRequest("POST", config.PushgatewayUrl, bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Push error creating request: %v\n", err)
		return
	}

	if (config.PushgatewayUsername != "") && (config.PushgatewayPassword != "") {
		req.SetBasicAuth(config.PushgatewayUsername, config.PushgatewayPassword)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Push error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Push response status: %s\n", resp.Status)
}
