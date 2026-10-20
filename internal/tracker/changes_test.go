package tracker

import "testing"

var lumberTycoon = Snapshot{
	Name:        "Lumber Tycoon 2",
	Description: "Welcome to Lumberland!",
	Updated:     "2026-10-20T01:03:25.971Z",
	Playing:     2443,
	IconURL:     "https://tr.rbxcdn.com/old/512/512/Image/Png",
}

func TestPublicChanges(t *testing.T) {
	t.Run("finds nothing when only the player count moved", func(t *testing.T) {
		current := lumberTycoon
		current.Playing = 9000

		if changes := PublicChanges(lumberTycoon, current); len(changes) != 0 {
			t.Fatalf("want no changes, got %v", changes)
		}
	})

	t.Run("lists every changed public field", func(t *testing.T) {
		current := lumberTycoon
		current.Name, current.Description, current.IconURL = "LT2", "New", "https://new"

		changes := PublicChanges(lumberTycoon, current)

		if len(changes) != 3 {
			t.Fatalf("want 3 changes, got %v", changes)
		}
	})
}

func TestSnapshotWithIconFrom(t *testing.T) {
	t.Run("keeps the last known icon when the new one is unknown", func(t *testing.T) {
		current := lumberTycoon
		current.IconURL = ""

		if got := current.WithIconFrom(lumberTycoon).IconURL; got != lumberTycoon.IconURL {
			t.Fatalf("want %q, got %q", lumberTycoon.IconURL, got)
		}
	})

	t.Run("prefers a fresh icon", func(t *testing.T) {
		current := lumberTycoon
		current.IconURL = "https://new"

		if got := current.WithIconFrom(lumberTycoon).IconURL; got != "https://new" {
			t.Fatalf("want fresh icon, got %q", got)
		}
	})
}
