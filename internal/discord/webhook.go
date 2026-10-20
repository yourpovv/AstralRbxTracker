// sends alerts through a discord webhook
package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/yourpovv/AstralRbxTracker/internal/tracker"
	"github.com/yourpovv/AstralRbxTracker/internal/wait"
)

const (
	username         = "Astral Update Tracker"
	defaultRetryWait = 2 * time.Second
)

// discord answered but didn't take the alert
type StatusError struct {
	Status int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("discord didn't accept the notification (%d)", e.Status)
}

type Webhook struct {
	URL  string
	HTTP *http.Client
}

var _ tracker.Notifier = Webhook{}

func (w Webhook) Notify(ctx context.Context, update tracker.Update) error {
	payload, err := json.Marshal(message{Username: username, Embeds: []embed{newEmbed(update)}})
	if err != nil {
		return err
	}

	response, err := w.post(ctx, payload)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusTooManyRequests {
		if err := wait.For(ctx, retryAfter(response.Header)); err != nil {
			return err
		}
		if response, err = w.post(ctx, payload); err != nil {
			return err
		}
	}

	if response.StatusCode/100 != 2 {
		return &StatusError{Status: response.StatusCode}
	}
	return nil
}

func (w Webhook) post(ctx context.Context, payload []byte) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("discord webhook URL is invalid: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := w.HTTP.Do(request)
	if err != nil {
		return nil, fmt.Errorf("can't reach Discord: %w", err)
	}
	response.Body.Close()
	return response, nil
}

// discord's Retry-After can be fractional seconds
func retryAfter(header http.Header) time.Duration {
	seconds, err := strconv.ParseFloat(header.Get("Retry-After"), 64)
	if err != nil {
		return defaultRetryWait
	}
	return time.Duration(seconds * float64(time.Second))
}
