package manage

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/yourpovv/AstralRbxTracker/internal/store"
	"github.com/yourpovv/AstralRbxTracker/internal/tracker"
)

type fixedPlaces map[int64]int64

func (f fixedPlaces) UniverseID(_ context.Context, placeID int64) (int64, error) {
	return f[placeID], nil
}

func managerIn(t *testing.T) Manager {
	dir := t.TempDir()
	return Manager{
		Games:  store.GameFile{Path: filepath.Join(dir, "games.json")},
		State:  store.StateFile{Path: filepath.Join(dir, "state.json")},
		Places: fixedPlaces{13822889: 2471084},
	}
}

func TestManagerAdd(t *testing.T) {
	t.Run("stores the universe ID it looked up", func(t *testing.T) {
		manager := managerIn(t)

		if err := manager.Add(context.Background(), 13822889, ""); err != nil {
			t.Fatal(err)
		}
		games, _ := manager.Games.Games()

		if len(games) != 1 || games[0].UniverseID != 2471084 {
			t.Fatalf("want one game with universe 2471084, got %v", games)
		}
	})

	t.Run("refuses a place that is already tracked", func(t *testing.T) {
		manager := managerIn(t)
		manager.Add(context.Background(), 13822889, "")

		err := manager.Add(context.Background(), 13822889, "")

		if !errors.Is(err, ErrAlreadyTracked) {
			t.Fatalf("want ErrAlreadyTracked, got %v", err)
		}
	})
}

func TestManagerRemove(t *testing.T) {
	t.Run("forgets the game's state", func(t *testing.T) {
		manager := managerIn(t)
		manager.Add(context.Background(), 13822889, "")
		manager.State.Save(map[int64]tracker.Snapshot{2471084: {Updated: "2026"}})

		if err := manager.Remove(13822889); err != nil {
			t.Fatal(err)
		}
		states, _ := manager.State.Load()

		if len(states) != 0 {
			t.Fatalf("want state cleared, got %v", states)
		}
	})

	t.Run("refuses a place that isn't tracked", func(t *testing.T) {
		if err := managerIn(t).Remove(5); !errors.Is(err, ErrNotTracked) {
			t.Fatalf("want ErrNotTracked, got %v", err)
		}
	})
}

func TestManagerDisable(t *testing.T) {
	t.Run("pauses the game", func(t *testing.T) {
		manager := managerIn(t)
		manager.Add(context.Background(), 13822889, "")

		if err := manager.Disable(13822889); err != nil {
			t.Fatal(err)
		}
		games, _ := manager.Games.Games()

		if games[0].Enabled {
			t.Fatal("want game disabled")
		}
	})
}
