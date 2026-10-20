package discord

import (
	"testing"
	"time"

	"github.com/yourpovv/AstralRbxTracker/internal/tracker"
)

var before = tracker.Snapshot{Name: "Lumber Tycoon 2", Updated: "2026-10-20T01:03:25.971Z", Playing: 2443}

func updateTo(current tracker.Snapshot) tracker.Update {
	return tracker.Update{
		Game:       tracker.Game{PlaceID: 13822889},
		Previous:   before,
		Current:    current,
		Changes:    tracker.PublicChanges(before, current),
		DetectedAt: time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC),
	}
}

func TestNewEmbed(t *testing.T) {
	t.Run("marks an unexplained update as a likely script or map change", func(t *testing.T) {
		current := before
		current.Updated = "2026-10-20T01:05:48.945Z"

		got := newEmbed(updateTo(current))

		if got.Color != updateColor || got.Fields[1].Value != noPublicChanges || len(got.Fields) != 4 {
			t.Fatalf("want update embed with timestamps and inferred note, got %+v", got)
		}
	})

	t.Run("reports a details change without timestamps", func(t *testing.T) {
		current := before
		current.Description = "New"

		got := newEmbed(updateTo(current))

		if got.Color != detailsColor || len(got.Fields) != 2 {
			t.Fatalf("want details embed without timestamps, got %+v", got)
		}
	})

	t.Run("shows the game icon as the thumbnail", func(t *testing.T) {
		current := before
		current.Updated, current.IconURL = "2026-10-20T01:05:48.945Z", "https://icon"

		got := newEmbed(updateTo(current))

		if got.Thumbnail == nil || got.Thumbnail.URL != "https://icon" {
			t.Fatalf("want icon thumbnail, got %+v", got.Thumbnail)
		}
	})
}

func TestWithThousands(t *testing.T) {
	t.Run("groups digits with commas", func(t *testing.T) {
		if got := withThousands(1337697364); got != "1,337,697,364" {
			t.Fatalf("got %q", got)
		}
	})
}
