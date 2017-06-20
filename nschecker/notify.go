package nschecker

import (
	"fmt"
	"log"
	"sync"

	"net/http"
	"net/url"

	"strings"

	"github.com/nlopes/slack"
)

// Notifier interface is construct an response.
type Notifier interface {
	Notify(state State, s Source) error
}

// SlackNotifier struct is construct an slack message.
type SlackNotifier struct {
	Cli *slack.Client

	channel string

	// url -> current state
	statesMu sync.Mutex
	states   map[string]State
}

func NewSlackNotifier(cli *slack.Client, channel string) Notifier {
	return &SlackNotifier{
		Cli:     cli,
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
	params := slack.PostMessageParameters{EscapeText: false}
	_, _, err := n.Cli.PostMessage(n.channel, msg, params)
	return err
}

// LineNotifier struct is construct an LINE message.
type LineNotifier struct {
	hc  *http.Client
	tok string

	statesMu sync.Mutex
	states   map[string]State
}

func NewLineNotifier(hc *http.Client, token string) Notifier {
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
