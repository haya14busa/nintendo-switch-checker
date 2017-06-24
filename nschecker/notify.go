package nschecker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"net/http"
	"net/url"

	"strings"
)

// Notifier is interface of notification
type Notifier interface {
	Notify(state State, s Source) error
	SendMessage(msg string) error
}

// SlackNotifier handles notification to slack
type SlackNotifier struct {
	hc      *http.Client
	tok     string
	channel string

	// url -> current state
	statesMu sync.Mutex
	states   map[string]State
}

func NewSlackNotifier(hc *http.Client, tok string, channel string) *SlackNotifier {
	return &SlackNotifier{
		hc:      hc,
		tok:     tok,
		channel: channel,
		states:  make(map[string]State),
	}
}

func (n *SlackNotifier) Notify(state State, s Source) error {
	defer func() {
		n.statesMu.Lock()
		n.states[s.URL] = state
		n.statesMu.Unlock()
	}()
	n.statesMu.Lock()
	oldState, ok := n.states[s.URL]
	n.statesMu.Unlock()
	if !ok && state == SOLDOUT {
		return nil
	}
	if oldState == state {
		log.Printf("same state: %v url=%v name=%v", state, s.URL, s.Name)
		return nil
	}
	channel := ""
	if state == AVAILABLE {
		channel = "<!channel|channel> "
	}
	msg := fmt.Sprintf("%s%v: %v (%v)", channel, state, s.URL, s.Name)
	return n.SendMessage(msg)
}

func (n *SlackNotifier) SendMessage(msg string) error {
	v := url.Values{}
	v.Set("token", n.tok)
	v.Set("channel", n.channel)
	v.Set("text", msg)

	r, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}

	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := n.hc.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return err
}

// LineNotifier handles notification to LINE.
type LineNotifier struct {
	hc  *http.Client
	tok string

	statesMu sync.Mutex
	states   map[string]State
}

func NewLineNotifier(hc *http.Client, token string) *LineNotifier {
	return &LineNotifier{
		hc:     hc,
		tok:    token,
		states: make(map[string]State),
	}
}

func (n *LineNotifier) Notify(state State, s Source) error {
	defer func() {
		n.statesMu.Lock()
		n.states[s.URL] = state
		n.statesMu.Unlock()
	}()
	n.statesMu.Lock()
	oldState, ok := n.states[s.URL]
	n.statesMu.Unlock()
	if !ok && state == SOLDOUT {
		return nil
	}
	if oldState == state {
		log.Printf("same state: %v url=%v name=%v", state, s.URL, s.Name)
		return nil
	}
	msg := fmt.Sprintf("%v: %v (%v)", state, s.URL, s.Name)
	return n.SendMessage(msg)
}

func (n *LineNotifier) SendMessage(msg string) error {
	v := url.Values{}
	v.Set("message", msg)

	r, err := http.NewRequest("POST", "https://notify-api.line.me/api/notify", strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}

	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Authorization", "Bearer "+n.tok)

	res, err := n.hc.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

// SlackWebhookNotifier handles notification to slack incoming webhook.
type SlackWebhookNotifier struct {
	hc      *http.Client
	url     string
	channel string

	// url -> current state
	statesMu sync.Mutex
	states   map[string]State
}

func NewSlackWebhookNotifier(hc *http.Client, url, channel string) *SlackWebhookNotifier {
	return &SlackWebhookNotifier{
		hc:      hc,
		url:     url,
		channel: channel,
		states:  make(map[string]State),
	}
}

type slackWebhookRequest struct {
	Channel  string `json:"channel"`
	Username string `json:"username"`
	Text     string `json:"text"`
}

func (n *SlackWebhookNotifier) Notify(state State, s Source) error {
	defer func() {
		n.statesMu.Lock()
		n.states[s.URL] = state
		n.statesMu.Unlock()
	}()
	n.statesMu.Lock()
	oldState, ok := n.states[s.URL]
	n.statesMu.Unlock()
	if !ok && state == SOLDOUT {
		return nil
	}
	if oldState == state {
		log.Printf("same state: %v url=%v name=%v", state, s.URL, s.Name)
		return nil
	}
	channel := ""
	if state == AVAILABLE {
		channel = "<!channel|channel> "
	}
	msg := fmt.Sprintf("%s%v: %v (%v)", channel, state, s.URL, s.Name)
	return n.SendMessage(msg)
}

func (n *SlackWebhookNotifier) SendMessage(msg string) error {
	req := &slackWebhookRequest{
		Channel:  n.channel,
		Username: "switch-checker",
		Text:     msg,
	}
	bs, err := json.Marshal(req)
	if err != nil {
		return err
	}
	r, err := http.NewRequest("POST", n.url, bytes.NewReader(bs))
	if err != nil {
		return err
	}

	r.Header.Set("Content-Type", "application/json")

	res, err := n.hc.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return err
}
