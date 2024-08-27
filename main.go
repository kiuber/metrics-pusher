package main

import (
	"bytes"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	MetricsUrl         string `validate:"required" arg:"env:METRICS_URL" help:"metrics url" default:""`
	PushgatewayUrl     string `validate:"required" arg:"env:PG_URL" help:"push gateway url" default:""`
	PushgatewayCrontab string `arg:"env:PG_CRONTAB" help:"push gateway crontab, default every 15 seconds" default:"*/15 * * * * *"`
}

func main() {
	config := &Config{}
	arg.MustParse(config)

	if config.MetricsUrl != "" && config.PushgatewayUrl != "" {
		c := cron.New(cron.WithSeconds(), cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)), cron.WithLogger(
			cron.VerbosePrintfLogger(log.New(os.Stdout, "crontab: ", log.LstdFlags))), cron.WithLocation(time.UTC))

		log.Printf("pushgateway crontab spec: %s", config.PushgatewayCrontab)
		c.AddFunc(config.PushgatewayCrontab, func() {
			log.Printf("Prepare push to %s", config.PushgatewayUrl)

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

			push(config.PushgatewayUrl, body)
		})
		c.Start()
	}

	log.Fatal(http.ListenAndServe(":9090", nil))
}

func push(url string, data []byte) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %s\n", resp.Status)
}
