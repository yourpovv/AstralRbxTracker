package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourpovv/AstralRbxTracker/internal/tracker"
)

func TestGameFileGames(t *testing.T) {
	t.Run("treats a missing file as no games", func(t *testing.T) {
		games, err := GameFile{Path: filepath.Join(t.TempDir(), "games.json")}.Games()

		if err != nil || len(games) != 0 {
			t.Fatalf("want empty list, got %v (%v)", games, err)
		}
	})

	t.Run("enables a hand-written game that leaves enabled out", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "games.json")
		os.WriteFile(path, []byte(`{"games":[{"place_id":1,"universe_id":2}]}`), 0o644)

		games, err := GameFile{Path: path}.Games()

		if err != nil || !games[0].Enabled {
			t.Fatalf("want enabled game, got %v (%v)", games, err)
		}
	})

	t.Run("rejects a game without a universe ID", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "games.json")
		os.WriteFile(path, []byte(`{"games":[{"place_id":1}]}`), 0o644)

		if _, err := (GameFile{Path: path}).Games(); err == nil {
			t.Fatal("want an error")
		}
	})
}

func TestStateFile(t *testing.T) {
	t.Run("reads back what it saved", func(t *testing.T) {
		file := StateFile{Path: filepath.Join(t.TempDir(), "state.json")}
		saved := map[int64]tracker.Snapshot{2471084: {Name: "LT2", Updated: "2026"}}

		if err := file.Save(saved); err != nil {
			t.Fatal(err)
		}
		loaded, err := file.Load()

		if err != nil || loaded[2471084] != saved[2471084] {
			t.Fatalf("want %v, got %v (%v)", saved, loaded, err)
		}
	})
}
