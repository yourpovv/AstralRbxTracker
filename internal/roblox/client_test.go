package roblox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

const lumberGameJSON = `{"data":[{"id":2471084,"name":"Lumber Tycoon 2","description":"Hi",
"updated":"2026-10-20T01:05:48.945Z","playing":3251}]}`

const lumberIconJSON = `{"data":[{"targetId":2471084,"state":"Completed","imageUrl":"https://icon"}]}`

func clientFor(t *testing.T, handler http.HandlerFunc) *Client {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Client{
		HTTP: server.Client(),
		Endpoints: Endpoints{
			Universes:  server.URL + "/universes/%d",
			Games:      server.URL + "/games",
			Thumbnails: server.URL + "/icons",
		},
		Retry: Retry{Attempts: 3, BaseWait: time.Millisecond, MaxWait: time.Millisecond},
		Log:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func TestClientUniverseID(t *testing.T) {
	t.Run("returns the universe for a real place", func(t *testing.T) {
		client := clientFor(t, func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"universeId":2471084}`)
		})

		universeID, err := client.UniverseID(context.Background(), 13822889)

		if err != nil || universeID != 2471084 {
			t.Fatalf("want 2471084, got %d (%v)", universeID, err)
		}
	})

	t.Run("rejects a place Roblox doesn't know", func(t *testing.T) {
		client := clientFor(t, func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"universeId":null}`)
		})

		_, err := client.UniverseID(context.Background(), 1)

		if !errors.Is(err, ErrPlaceNotFound) {
			t.Fatalf("want ErrPlaceNotFound, got %v", err)
		}
	})
}

func TestClientSnapshots(t *testing.T) {
	t.Run("retries after a rate limit", func(t *testing.T) {
		var calls atomic.Int32
		client := clientFor(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/games" && calls.Add(1) == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			serveLumber(w, r)
		})

		snapshots, err := client.Snapshots(context.Background(), []int64{2471084})

		if err != nil || snapshots[2471084].Updated == "" {
			t.Fatalf("want snapshot after retry, got %v (%v)", snapshots, err)
		}
	})

	t.Run("gives up after the last attempt", func(t *testing.T) {
		client := clientFor(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		})

		_, err := client.Snapshots(context.Background(), []int64{2471084})

		var statusErr *StatusError
		if !errors.As(err, &statusErr) || statusErr.Status != http.StatusServiceUnavailable {
			t.Fatalf("want 503 StatusError, got %v", err)
		}
	})

	t.Run("still returns games when icons are down", func(t *testing.T) {
		client := clientFor(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/icons" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			serveLumber(w, r)
		})

		snapshots, err := client.Snapshots(context.Background(), []int64{2471084})

		if err != nil || snapshots[2471084].IconURL != "" {
			t.Fatalf("want snapshot without icon, got %v (%v)", snapshots, err)
		}
	})

	t.Run("attaches the game icon", func(t *testing.T) {
		client := clientFor(t, serveLumber)

		snapshots, err := client.Snapshots(context.Background(), []int64{2471084})

		if err != nil || snapshots[2471084].IconURL != "https://icon" {
			t.Fatalf("want icon attached, got %v (%v)", snapshots, err)
		}
	})
}

func TestBatches(t *testing.T) {
	t.Run("splits into groups of 50", func(t *testing.T) {
		ids := make([]int64, 120)

		got := batches(ids)

		if len(got) != 3 || len(got[2]) != 20 {
			t.Fatalf("want 50+50+20, got %d batches", len(got))
		}
	})
}

func serveLumber(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/icons" {
		fmt.Fprint(w, lumberIconJSON)
		return
	}
	fmt.Fprint(w, lumberGameJSON)
}
