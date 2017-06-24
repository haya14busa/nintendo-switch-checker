package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/haya14busa/nintendo-switch-checker/nschecker"
)

var (
	interval = flag.Duration("interval", 1*time.Minute, "Check interval")
	channel  = flag.String("channel", "", "Slack channel name where checker posts comments")
	once     = flag.Bool("once", false, "Check once")
	notifier = flag.String("notifier", "", "Notifier target, slack or line")
)

const (
	debugNotify = true
)

const usageMessage = "" +
	`Usage:	nintendo-switch-checker [flags]

	Set notification token by environment variable.

	export SLACK_API_TOKEN=<SLACK_API_TOKEN>
	# or
	export LINE_NOTIFY_TOKEN=<LINE_NOTIFY_TOKEN>
	# or
	export SLACK_WEBHOOK_URL=<SLACK_WEBHOOK_URL>

	### How to get LINE_NOTIFY_TOKEN
	1. Go to https://notify-bot.line.me/ and LOGIN
	3. Go to my page https://notify-bot.line.me/my/
	4. Click "Generate token" and input token name and select target.
	5. $ LINE_NOTIFY_TOKEN=xxxxxx nintendo-switch-checker -notifier=line
`

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usageMessage)
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *notifier == "" {
		log.Println("Please set -notifier=slack (or line)")
		return
	}

	var n nschecker.Notifier
	switch *notifier {
	default:
		log.Printf("Not support notifier: %s", *notifier)
		return
	case "slack":
		tok := os.Getenv("SLACK_API_TOKEN")
		if tok == "" {
			log.Println("Please set environment variable SLACK_API_TOKEN")
			return
		}
		if *channel == "" {
			log.Println("Please set -slack-channel flag")
			return
		}

		n = nschecker.NewSlackNotifier(http.DefaultClient, tok, *channel)
	case "line":
		tok := os.Getenv("LINE_NOTIFY_TOKEN")
		if tok == "" {
			log.Println("Please set environment variable LINE_NOTIFY_TOKEN")
			return
		}
		lineNotifier := nschecker.NewLineNotifier(http.DefaultClient, tok)
		lineNotifier.SendMessage("Switch checker started")
		n = lineNotifier
	case "slack-webhook":
		url := os.Getenv("SLACK_WEBHOOK_URL")
		if url == "" {
			log.Println("Please set environment variable SLACK_WEBHOOK_URL")
			return
		}

		if *channel == "" {
			log.Println("Please set -channel flag")
			return
		}

		n = nschecker.NewSlackWebhookNotifier(http.DefaultClient, url, *channel)
	}

	c := &Checker{
		Notifier: n,
		Interval: *interval,
		Once:     *once,
	}
	if err := c.run(); err != nil {
		log.Fatal(err)
	}
}

type Checker struct {
	Notifier nschecker.Notifier
	Interval time.Duration
	Once     bool
}

func (c *Checker) run() error {
	if c.Once {
		c.runChecks()
		return nil
	}
	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()
	c.runChecks()
	for range ticker.C {
		c.runChecks()
	}
	return nil
}

func (c *Checker) runChecks() {
	log.Println("Run checkers")
	var wg sync.WaitGroup
	for _, s := range nschecker.Sources {
		wg.Add(1)
		go func(s nschecker.Source) {
			defer wg.Done()
			c.check(s)
		}(s)
	}
	wg.Wait()
}

func (c *Checker) check(s nschecker.Source) {
	state, err := nschecker.Check(s, nil)
	if err != nil {
		log.Printf("Check failed: %s: %v", s.Name, err)
	}
	log.Printf("%v: %v (%s)", state, s.URL, s.Name)
	if err := c.Notifier.Notify(state, s); err != nil {
		log.Printf("fail to notify: %v", err)
	}
}
