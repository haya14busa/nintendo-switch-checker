package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/haya14busa/nintendo-switch-checker/nschecker"
	"github.com/nlopes/slack"
)

var (
	interval = flag.Duration("interval", 1*time.Minute, "Check interval")
	channel  = flag.String("channel", "", "Slack channel name where checker posts comments")
	once     = flag.Bool("once", false, "Check once")
)

var sources = []nschecker.Source{
	{
		Name:          "Amazon - Nintendo Switch Joy-Con (L) / (R) グレー",
		URL:           "https://www.amazon.co.jp/%E4%BB%BB%E5%A4%A9%E5%A0%82-Nintendo-Switch-Joy-Con-%E3%82%B0%E3%83%AC%E3%83%BC/dp/B01N5QLLT3/",
		AvailableText: `この商品は、<a href="/gp/help/customer/display.html?ie=UTF8&amp;nodeId=643004">Amazon.co.jp</a> が販売、発送します。`,
	},
	{
		Name:          "Amazon - Nintendo Switch Joy-Con (L) ネオンブルー/ (R) ネオンレッド",
		URL:           "https://www.amazon.co.jp/Nintendo-Switch-Joy-Con-%E3%83%8D%E3%82%AA%E3%83%B3%E3%83%96%E3%83%AB%E3%83%BC-%E3%83%8D%E3%82%AA%E3%83%B3%E3%83%AC%E3%83%83%E3%83%89/dp/B01NCXFWIZ/",
		AvailableText: `この商品は、<a href="/gp/help/customer/display.html?ie=UTF8&amp;nodeId=643004">Amazon.co.jp</a> が販売、発送します。`,
	},
	{
		Name:        "My Nintendo Store",
		URL:         "https://store.nintendo.co.jp/customize.html",
		SoldOutText: `<button class="btn btn__primary_soldout to_cart" type="submit"><span>SOLD OUT</span></button>`,
	},
	{
		Name:        "Yodobashi - Nintendo Switch Joy-Con(L)/(R)グレー [Nintendo Switch本体]",
		URL:         "http://www.yodobashi.com/product/100000001003431565/",
		SoldOutText: `<div class="salesInfo"><p>予定数の販売を終了しました</p></div>`,
	},
	{
		Name:        "Yodobashi - Nintendo Switch Joy-Con(L)ネオンブルー/(R)ネオンレッド [Nintendo Switch本体]",
		URL:         "http://www.yodobashi.com/product/100000001003431566/",
		SoldOutText: `<div class="salesInfo"><p>予定数の販売を終了しました</p></div>`,
	},
	{
		Name:        "Joshin - Nintendo Switch 本体【Joy-Con(L)/(R) グレー】",
		URL:         "http://joshinweb.jp/game/40519/4902370535709.html",
		SoldOutText: `<span class="fsL"><font color="blue"><b>販売休止中です</b></font><br></span>`,
	},
	{
		Name:        "Joshin - Nintendo Switch 本体【Joy-Con(L) ネオンブルー/(R) ネオンレッド】",
		URL:         "http://joshinweb.jp/game/40519/4902370535716.html",
		SoldOutText: `<span class="fsL"><font color="blue"><b>販売休止中です</b></font><br></span>`,
	},
}

const usageMessage = "" +
	`Usage:	nintendo-switch-checker [flags]

	export SLACK_API_TOKEN=<SLACK_API_TOKEN>
`

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usageMessage)
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
	}
	flag.Parse()
	token := os.Getenv("SLACK_API_TOKEN")
	if token == "" {
		log.Println("Please set environment variable SLACK_API_TOKEN")
		return
	}
	if *channel == "" {
		log.Println("Please set -channel flag")
		return
	}

	c := &Checker{
		Notifier: NewNotifier(slack.New(token), *channel),
		Interval: *interval,
		Once:     *once,
	}
	if err := c.run(); err != nil {
		log.Fatal(err)
	}
}

type Checker struct {
	Notifier *Notifier
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
	for _, s := range sources {
		wg.Add(1)
		go func(s nschecker.Source) {
			defer wg.Done()
			c.check(s)
		}(s)
	}
	wg.Wait()
}

func (c *Checker) check(s nschecker.Source) {
	state, err := nschecker.Check(s)
	if err != nil {
		log.Printf("Check failed: %s: %v", s.Name, err)
	}
	log.Printf("%v: %v (%s)", state, s.URL, s.Name)
	if err := c.Notifier.Notify(state, s); err != nil {
		log.Printf("fail to notify: %v", err)
	}
}

type Notifier struct {
	Cli *slack.Client

	channel string

	// url -> current state
	statesMu sync.Mutex
	states   map[string]nschecker.State
}

func NewNotifier(cli *slack.Client, channel string) *Notifier {
	return &Notifier{
		Cli:     cli,
		channel: channel,
		states:  make(map[string]nschecker.State),
	}
}

func (n *Notifier) Notify(state nschecker.State, s nschecker.Source) error {
	defer func() {
		n.statesMu.Lock()
		n.states[s.URL] = state
		n.statesMu.Unlock()
	}()
	n.statesMu.Lock()
	oldState := n.states[s.URL]
	n.statesMu.Unlock()

	if oldState == state {
		log.Printf("same state: %v url=%v name=%v", state, s.URL, s.Name)
		return nil
	}
	msg := fmt.Sprintf("%v: %v (%v)", state, s.URL, s.Name)
	params := slack.PostMessageParameters{EscapeText: false}
	_, _, err := n.Cli.PostMessage(n.channel, msg, params)
	return err
}
