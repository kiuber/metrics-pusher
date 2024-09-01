package main

import (
	"github.com/alexflint/go-arg"
	"github.com/kiuber/metrics-pusher/mper"
	"log"
	"net/http"
)

func main() {
	config := &mper.PullPushConfig{}
	arg.MustParse(config)

	mper.PullPushCrontab(*config)

	log.Fatal(http.ListenAndServe(":9090", nil))
}
