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

const (
	debugNotify = true
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
	{
		Name:        "omni7(7net) - Nintendo Switch Joy-Con (L) / (R) グレー",
		URL:         "http://7net.omni7.jp/detail/2110595636",
		SoldOutText: `<input class="linkBtn js-pressTwice" type="submit" value="在庫切れ" title="在庫切れ"`,
	},
	{
		Name:        "omni7(7net) - Nintendo Switch Joy-Con (L) ネオンブルー/ (R) ネオンレッド",
		URL:         "http://7net.omni7.jp/detail/2110595637",
		SoldOutText: `<input class="linkBtn js-pressTwice" type="submit" value="在庫切れ" title="在庫切れ"`,
	},
	{
		Name:        "omni7(iyec) - Nintendo Switch Joy-Con (L) / (R) グレー",
		URL:         "http://iyec.omni7.jp/detail/4902370535709",
		SoldOutText: `<input class="linkBtn js-pressTwice" type="submit" value="在庫切れ" title="在庫切れ"`,
	},
	{
		Name:        "omni7(iyec) - Nintendo Switch Joy-Con (L) ネオンブルー/ (R) ネオンレッド",
		URL:         "http://iyec.omni7.jp/detail/4902370535716",
		SoldOutText: `<input class="linkBtn js-pressTwice" type="submit" value="在庫切れ" title="在庫切れ"`,
	},
	{
		Name:        "nojima - Nintendo Switch Joy-Con (L) / (R) グレー",
		URL:         "https://online.nojima.co.jp/Nintendo-HAC-S-KAAAA-%E3%80%90NSW%E3%80%91-%E3%83%8B%E3%83%B3%E3%83%86%E3%83%B3%E3%83%89%E3%83%BC%E3%82%B9%E3%82%A4%E3%83%83%E3%83%81%E6%9C%AC%E4%BD%93-Joy-Con%28L%29-%28R%29-%E3%82%B0%E3%83%AC%E3%83%BC/4902370535709/1/cd/",
		SoldOutText: `<span>完売御礼</span>`,
	},
	{
		Name:        "nojima - Nintendo Switch Joy-Con (L) ネオンブルー/ (R) ネオンレッド",
		URL:         "https://online.nojima.co.jp/Nintendo-HAC-S-KABAA-%E3%80%90NSW%E3%80%91-%E3%83%8B%E3%83%B3%E3%83%86%E3%83%B3%E3%83%89%E3%83%BC%E3%82%B9%E3%82%A4%E3%83%83%E3%83%81%E6%9C%AC%E4%BD%93-Joy-Con%28L%29-%E3%83%8D%E3%82%AA%E3%83%B3%E3%83%96%E3%83%AB%E3%83%BC-%28R%29-%E3%83%8D%E3%82%AA%E3%83%B3%E3%83%AC%E3%83%83%E3%83%89/4902370535716/1/cd/",
		SoldOutText: `<span>完売御礼</span>`,
	},
	{
		Name:        "yamada - Nintendo Switch Joy-Con (L) / (R) グレー",
		URL:         "http://www.yamada-denkiweb.com/1177991016",
		SoldOutText: `<button type="submit" class="btn btn-disabled btn-block" disabled="disabled">売り切れました</button>`,
	},
	{
		Name:        "yamada - Nintendo Switch Joy-Con (L) ネオンブルー/ (R) ネオンレッド",
		URL:         "http://www.yamada-denkiweb.com/1177992013",
		SoldOutText: `<button type="submit" class="btn btn-disabled btn-block" disabled="disabled">売り切れました</button>`,
	},
	{
		Name:        "toysrus - Nintendo Switch Joy-Con (L) / (R) グレー",
		URL:         "https://www.toysrus.co.jp/s/dsg-572182200",
		SoldOutText: `<span id="isStock_c" >在庫なし/入荷予定あり</span>`,
	},
	{
		Name:        "toysrus - Nintendo Switch Joy-Con (L) ネオンブルー/ (R) ネオンレッド",
		URL:         "https://www.toysrus.co.jp/s/dsg-572186500",
		SoldOutText: `<span id="isStock_c" >在庫なし/入荷予定あり</span>`,
	},
	{
		Name:        "tsutaya - Nintendo Switch Joy-Con (L) / (R) グレー",
		URL:         "http://shop.tsutaya.co.jp/Nintendo-Switch-Joy-Con-L-R-%E3%82%B0%E3%83%AC%E3%83%BC-HACSKAAAA/product-game-4902370535709/",
		SoldOutText: `<img src="/library/img/base/ic/btn_nostockL.png" alt="在庫なし" />`,
	},
	{
		Name:        "tsutaya - Nintendo Switch Joy-Con (L) ネオンブルー/ (R) ネオンレッド",
		URL:         "http://shop.tsutaya.co.jp/Nintendo-Switch-Joy-Con-L-%E3%83%8D%E3%82%AA%E3%83%B3%E3%83%96%E3%83%AB%E3%83%BC-R-%E3%83%8D%E3%82%AA%E3%83%B3%E3%83%AC%E3%83%83%E3%83%89-HACSKABAA/product-game-4902370535716/",
		SoldOutText: `<img src="/library/img/base/ic/btn_nostockL.png" alt="在庫なし" />`,
	},
	{
		Name:        "sofmap - Nintendo Switch",
		URL:         "http://www.sofmap.com/topics/exec/?id=5500",
		SoldOutText: `<IMG src="/images/system_icon/zaiko06.gif" alt="在庫切れ" border="0">`,
	},
	{
		Name:        "rakuten - Nintendo Switch Joy-Con(L)/(R) グレー + ゼルダの伝説　ブレス オブ ザ ワイルド Nintendo Switch版",
		URL:         "http://books.rakuten.co.jp/rb/14779136/",
		SoldOutText: `<span class="status">ご注文できない商品*`,
	},
	{
		Name:        "rakuten - Nintendo Switch Joy-Con(L)/(R) グレー + マリオカート8 デラックス",
		URL:         "http://books.rakuten.co.jp/rb/14785337/",
		SoldOutText: `<span class="status">ご注文できない商品*`,
	},
	{
		Name:        "rakuten - Nintendo Switch Joy-Con(L)/(R) グレー 楽天あんしん延長保証",
		URL:         "http://books.rakuten.co.jp/rb/14655634/",
		SoldOutText: `<span class="status">ご注文できない商品*`,
	},
	{
		Name:        "rakuten - Nintendo Switch Joy-Con(L) ネオンブルー/(R) ネオンレッド 楽天あんしん延長保証",
		URL:         "http://books.rakuten.co.jp/rb/14655635/",
		SoldOutText: `<span class="status">ご注文できない商品*`,
	},
	{
		Name:        "rakuten - Nintendo Switch Joy-Con(L) ネオンブルー/(R) ネオンレッド + マリオカート8 デラックス",
		URL:         "http://books.rakuten.co.jp/rb/14787497/",
		SoldOutText: `<span class="status">ご注文できない商品*`,
	},
	{
		Name:        "rakuten - Nintendo Switch Joy-Con(L) ネオンブルー/(R) ネオンレッド + 1-2-Switch ",
		URL:         "http://books.rakuten.co.jp/rb/14779141/",
		SoldOutText: `<span class="status">ご注文できない商品*`,
	},
	{
		Name:        "rakuten - Nintendo Switch Joy-Con(L)/(R) グレー",
		URL:         "http://books.rakuten.co.jp/rb/14647221/",
		SoldOutText: `<span class="status">ご注文できない商品*`,
	},
	{
		Name:        "rakuten - Nintendo Switch Joy-Con(L) ネオンブルー/(R) ネオンレッド",
		URL:         "http://books.rakuten.co.jp/rb/14647222/",
		SoldOutText: `<span class="status">ご注文できない商品*`,
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
	oldState, ok := n.states[s.URL]
	n.statesMu.Unlock()
	if !ok && state == nschecker.SOLDOUT {
		return nil
	}
	if oldState == state {
		log.Printf("same state: %v url=%v name=%v", state, s.URL, s.Name)
		return nil
	}
	channel := ""
	if state == nschecker.AVAILABLE {
		channel = "<!channel|channel> "
	}
	msg := fmt.Sprintf("%s%v: %v (%v)", channel, state, s.URL, s.Name)
	params := slack.PostMessageParameters{EscapeText: false}
	if !debugNotify {
		return nil
	}
	_, _, err := n.Cli.PostMessage(n.channel, msg, params)
	return err
}
