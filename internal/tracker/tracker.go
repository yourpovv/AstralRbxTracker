package tracker

import (
	"context"
	"log/slog"
	"time"
)

type GameList interface {
	Games() ([]Game, error)
}

type StateStore interface {
	Load() (map[int64]Snapshot, error)
	Save(states map[int64]Snapshot) error
}

type Source interface {
	Snapshots(ctx context.Context, universeIDs []int64) (map[int64]Snapshot, error)
}

type Notifier interface {
	Notify(ctx context.Context, update Update) error
}

// alerts should pass before state is saved
type Tracker struct {
	Games    GameList
	State    StateStore
	Source   Source
	Notifier Notifier
	Log      *slog.Logger
	Now      func() time.Time
}

func (t *Tracker) Check(ctx context.Context) error {
	games, err := t.enabledGames()
	if err != nil {
		return err
	}
	if len(games) == 0 {
		t.Log.Info("no games enabled, add one: AstralRbxTracker add <place_id>")
		return nil
	}

	states, err := t.State.Load()
	if err != nil {
		return err
	}
	snapshots, err := t.Source.Snapshots(ctx, universeIDs(games))
	if err != nil {
		return err
	}

	for _, game := range games {
		if err := t.checkGame(ctx, game, snapshots, states); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tracker) checkGame(
	ctx context.Context, game Game, snapshots, states map[int64]Snapshot,
) error {
	current, found := snapshots[game.UniverseID]
	if !found {
		t.Log.Warn("Roblox sent no data, skipped", "place", game.PlaceID)
		return nil
	}

	if previous, seen := states[game.UniverseID]; seen {
		current = current.WithIconFrom(previous)
		if err := t.announce(ctx, t.updateFor(game, previous, current)); err != nil {
			return err
		}
	} else {
		t.Log.Info("baseline saved", "game", current.Name, "updated", current.Updated)
	}

	states[game.UniverseID] = current
	return t.State.Save(states)
}

func (t *Tracker) updateFor(game Game, previous, current Snapshot) Update {
	return Update{
		Game:       game,
		Previous:   previous,
		Current:    current,
		Changes:    PublicChanges(previous, current),
		DetectedAt: t.Now(),
	}
}

func (t *Tracker) announce(ctx context.Context, update Update) error {
	name := update.DisplayName()
	if !update.IsNews() {
		t.Log.Info("no update", "game", name, "updated", update.Current.Updated)
		return nil
	}
	if err := t.Notifier.Notify(ctx, update); err != nil {
		return err
	}
	t.Log.Info("notified", "game", name,
		"from", update.Previous.Updated, "to", update.Current.Updated,
		"changes", len(update.Changes))
	return nil
}

func (t *Tracker) enabledGames() ([]Game, error) {
	games, err := t.Games.Games()
	if err != nil {
		return nil, err
	}
	enabled := []Game{}
	for _, game := range games {
		if game.Enabled {
			enabled = append(enabled, game)
		}
	}
	return enabled, nil
}

func universeIDs(games []Game) []int64 {
	ids := make([]int64, 0, len(games))
	for _, game := range games {
		ids = append(ids, game.UniverseID)
	}
	return ids
}
