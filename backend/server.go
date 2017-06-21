package backend

import (
	"errors"
	"net/http"
	"os"
	"sync"

	"github.com/haya14busa/nintendo-switch-checker/nschecker"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	hc := urlfetch.Client(ctx)
	var wg sync.WaitGroup
	for _, s := range nschecker.Sources {
		wg.Add(1)
		go func(s nschecker.Source) {
			defer wg.Done()
			err := check(ctx, hc, s)
			if err != nil {
				log.Errorf(ctx, "Check failed: %s: %v", s.Name, err)
				return
			}
		}(s)
	}
	wg.Wait()
}

func check(ctx context.Context, hc *http.Client, s nschecker.Source) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	n, err := newNotifier(ctx, hc)
	if err != nil {
		return err
	}

	state, err := nschecker.Check(s, nschecker.HTTPClient(hc))
	if err != nil {
		return err
	}
	log.Infof(ctx, "%v: %v (%s)", state, s.URL, s.Name)
	return n.Notify(state, s)
}

func newNotifier(ctx context.Context, hc *http.Client) (nschecker.Notifier, error) {
	tok := os.Getenv("SLACK_API_TOKEN")
	if tok != "" {
		channel := os.Getenv("SLACK_CHANNEL")
		if channel == "" {
			return nil, errors.New("Please set enviroment variable SLACK_CHANNEL")
		}

		return nschecker.NewSlackNotifier(hc, tok, channel), nil
	}

	tok = os.Getenv("LINE_NOTIFY_TOKEN")
	if tok != "" {
		return nschecker.NewLineNotifier(hc, tok), nil
	}

	url := os.Getenv("SLACK_WEBHOOK_URL")
	if url != "" {
		channel := os.Getenv("SLACK_CHANNEL")
		if channel == "" {
			return nil, errors.New("Please set enviroment variable SLACK_CHANNEL")
		}
		return nschecker.NewSlackWebhookNotifier(hc, url, channel), nil
	}

	return nil, errors.New("Not notify token")
}
