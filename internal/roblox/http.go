package roblox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yourpovv/AstralRbxTracker/internal/wait"
)

const maxResponseBytes = 4 << 20

// backs off on 429s and outages
type Retry struct {
	Attempts int
	BaseWait time.Duration
	MaxWait  time.Duration
}

var DefaultRetry = Retry{Attempts: 4, BaseWait: 2 * time.Second, MaxWait: time.Minute}

// roblox didn't give us anything we can use
type StatusError struct {
	URL    string
	Status int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("roblox refused the request (%d): %s", e.Status, e.URL)
}

type retryableError struct {
	err        error
	retryAfter time.Duration
}

func (e *retryableError) Error() string { return e.err.Error() }
func (e *retryableError) Unwrap() error { return e.err }

func (c *Client) getJSON(ctx context.Context, target string, out any) error {
	body, err := c.getWithRetries(ctx, target)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("roblox sent back something unexpected: %w", err)
	}
	return nil
}

func (c *Client) getWithRetries(ctx context.Context, target string) ([]byte, error) {
	for attempt := 1; attempt < c.Retry.Attempts; attempt++ {
		body, err := c.get(ctx, target)
		var retryable *retryableError
		if !errors.As(err, &retryable) {
			return body, err
		}

		pause := c.Retry.pause(attempt, retryable.retryAfter)
		c.Log.Warn("roblox busy, retrying", "err", retryable.err, "in", pause)
		if err := wait.For(ctx, pause); err != nil {
			return nil, err
		}
	}
	return c.get(ctx, target)
}

func (c *Client) get(ctx context.Context, target string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	response, err := c.HTTP.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, &retryableError{err: fmt.Errorf("can't reach Roblox: %w", err)}
	}
	defer response.Body.Close()

	statusErr := &StatusError{URL: target, Status: response.StatusCode}
	if isRetryable(response.StatusCode) {
		return nil, &retryableError{err: statusErr, retryAfter: retryAfter(response.Header)}
	}
	if response.StatusCode != http.StatusOK {
		return nil, statusErr
	}
	return io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
}

func (r Retry) pause(attempt int, retryAfter time.Duration) time.Duration {
	backoff := r.BaseWait << (attempt - 1)
	return min(max(backoff, retryAfter), r.MaxWait)
}

func isRetryable(status int) bool {
	return status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

// x-ratelimit-reset can list a few windows like "43, 43"
func retryAfter(header http.Header) time.Duration {
	value := header.Get("Retry-After")
	if value == "" {
		value = header.Get("X-Ratelimit-Reset")
	}
	seconds, err := strconv.Atoi(strings.TrimSpace(strings.Split(value, ",")[0]))
	if err != nil {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
