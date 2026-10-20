package tracker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

const lumberUniverse = 2471084

var (
	lumberGame = Game{PlaceID: 13822889, UniverseID: lumberUniverse, Enabled: true}
	noon       = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	errDiscord = errors.New("discord down")
)

type fakeGames []Game

func (f fakeGames) Games() ([]Game, error) { return f, nil }

type fakeState struct{ saved map[int64]Snapshot }

func (f *fakeState) Load() (map[int64]Snapshot, error) {
	copied := map[int64]Snapshot{}
	for id, snapshot := range f.saved {
		copied[id] = snapshot
	}
	return copied, nil
}

func (f *fakeState) Save(states map[int64]Snapshot) error {
	f.saved = states
	return nil
}

type fakeSource map[int64]Snapshot

func (f fakeSource) Snapshots(context.Context, []int64) (map[int64]Snapshot, error) {
	return f, nil
}

type fakeNotifier struct {
	sent []Update
	err  error
}

func (f *fakeNotifier) Notify(_ context.Context, update Update) error {
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, update)
	return nil
}

func trackerWith(state *fakeState, live Snapshot, notifier *fakeNotifier) *Tracker {
	return &Tracker{
		Games:    fakeGames{lumberGame},
		State:    state,
		Source:   fakeSource{lumberUniverse: live},
		Notifier: notifier,
		Log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		Now:      func() time.Time { return noon },
	}
}

func published() Snapshot {
	live := lumberTycoon
	live.Updated = "2026-10-20T01:05:48.945Z"
	return live
}

func TestTrackerCheck(t *testing.T) {
	t.Run("saves a silent baseline the first time it sees a game", func(t *testing.T) {
		state, notifier := &fakeState{}, &fakeNotifier{}

		if err := trackerWith(state, lumberTycoon, notifier).Check(context.Background()); err != nil {
			t.Fatal(err)
		}

		if len(notifier.sent) != 0 || state.saved[lumberUniverse] != lumberTycoon {
			t.Fatalf("want baseline and no alert, got %d alerts", len(notifier.sent))
		}
	})

	t.Run("alerts once per new update", func(t *testing.T) {
		state := &fakeState{saved: map[int64]Snapshot{lumberUniverse: lumberTycoon}}
		notifier := &fakeNotifier{}
		tracker := trackerWith(state, published(), notifier)

		for range 2 {
			if err := tracker.Check(context.Background()); err != nil {
				t.Fatal(err)
			}
		}

		if len(notifier.sent) != 1 || !notifier.sent[0].Published() {
			t.Fatalf("want 1 published alert, got %d", len(notifier.sent))
		}
	})

	t.Run("keeps the old state when the alert fails so it retries", func(t *testing.T) {
		state := &fakeState{saved: map[int64]Snapshot{lumberUniverse: lumberTycoon}}

		err := trackerWith(state, published(), &fakeNotifier{err: errDiscord}).Check(context.Background())

		if !errors.Is(err, errDiscord) || state.saved[lumberUniverse] != lumberTycoon {
			t.Fatalf("want old state kept and error %v, got %v", errDiscord, err)
		}
	})

	t.Run("sends a test alert with live data and leaves state alone", func(t *testing.T) {
		state := &fakeState{saved: map[int64]Snapshot{lumberUniverse: lumberTycoon}}
		notifier := &fakeNotifier{}

		if err := trackerWith(state, published(), notifier).SendTest(context.Background(), lumberGame); err != nil {
			t.Fatal(err)
		}

		if len(notifier.sent) != 1 || !notifier.sent[0].Published() || notifier.sent[0].Current != published() {
			t.Fatalf("want 1 published alert with live data, got %+v", notifier.sent)
		}
		if state.saved[lumberUniverse] != lumberTycoon {
			t.Fatalf("want state untouched, got %+v", state.saved)
		}
	})

	t.Run("skips disabled games", func(t *testing.T) {
		state, notifier := &fakeState{}, &fakeNotifier{}
		tracker := trackerWith(state, lumberTycoon, notifier)
		disabled := lumberGame
		disabled.Enabled = false
		tracker.Games = fakeGames{disabled}

		if err := tracker.Check(context.Background()); err != nil {
			t.Fatal(err)
		}

		if len(state.saved) != 0 {
			t.Fatalf("want nothing saved, got %v", state.saved)
		}
	})
}
