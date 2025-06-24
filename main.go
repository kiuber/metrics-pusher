package main

import (
	"log"
	"net/http"

	"github.com/alexflint/go-arg"
	"github.com/kiuber/metrics-pusher/mper"
)

func main() {
	config := &mper.PullPushConfig{}
	arg.MustParse(config)

	mper.PullPushCrontab(*config)

	log.Fatal(http.ListenAndServe(":9090", nil))
}
