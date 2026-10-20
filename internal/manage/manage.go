// manage changes which games are tracked
package manage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/yourpovv/AstralRbxTracker/internal/store"
	"github.com/yourpovv/AstralRbxTracker/internal/tracker"
)

var (
	ErrAlreadyTracked = errors.New("already tracked")
	ErrNotTracked     = errors.New("isn't tracked, see: list")
)

type PlaceResolver interface {
	UniverseID(ctx context.Context, placeID int64) (int64, error)
}

type Manager struct {
	Games  store.GameFile
	State  store.StateFile
	Places PlaceResolver
}

func (m Manager) HasGames() (bool, error) {
	games, err := m.Games.Games()
	return len(games) > 0, err
}

// empty game name will use the published game name
func (m Manager) Add(ctx context.Context, placeID int64, name string) error {
	games, err := m.Games.Games()
	if err != nil {
		return err
	}
	if _, found := indexOf(games, placeID); found {
		return fmt.Errorf("place %d: %w", placeID, ErrAlreadyTracked)
	}

	universeID, err := m.Places.UniverseID(ctx, placeID)
	if err != nil {
		return err
	}
	game := tracker.Game{PlaceID: placeID, UniverseID: universeID, Name: name, Enabled: true}
	return m.Games.Save(append(games, game))
}

// re-adding a game starts a fresh baseline
func (m Manager) Remove(placeID int64) error {
	games, index, err := m.find(placeID)
	if err != nil {
		return err
	}
	removed := games[index]
	if err := m.Games.Save(append(games[:index], games[index+1:]...)); err != nil {
		return err
	}

	states, err := m.State.Load()
	if err != nil {
		return err
	}
	delete(states, removed.UniverseID)
	return m.State.Save(states)
}

func (m Manager) Enable(placeID int64) error {
	return m.setEnabled(placeID, true)
}

func (m Manager) Disable(placeID int64) error {
	return m.setEnabled(placeID, false)
}

func (m Manager) List(out io.Writer) error {
	games, err := m.Games.Games()
	if err != nil {
		return err
	}
	if len(games) == 0 {
		fmt.Fprintln(out, "No games tracked. Add one: add <place_id>")
		return nil
	}

	states, err := m.State.Load()
	if err != nil {
		return err
	}
	for _, game := range games {
		fmt.Fprintln(out, describe(game, states))
	}
	return nil
}

func (m Manager) setEnabled(placeID int64, enabled bool) error {
	games, index, err := m.find(placeID)
	if err != nil {
		return err
	}
	games[index].Enabled = enabled
	return m.Games.Save(games)
}

func (m Manager) find(placeID int64) ([]tracker.Game, int, error) {
	games, err := m.Games.Games()
	if err != nil {
		return nil, 0, err
	}
	index, found := indexOf(games, placeID)
	if !found {
		return nil, 0, fmt.Errorf("place %d: %w", placeID, ErrNotTracked)
	}
	return games, index, nil
}

func indexOf(games []tracker.Game, placeID int64) (int, bool) {
	for index, game := range games {
		if game.PlaceID == placeID {
			return index, true
		}
	}
	return 0, false
}

func describe(game tracker.Game, states map[int64]tracker.Snapshot) string {
	status := "off"
	if game.Enabled {
		status = "on "
	}
	name, updated := game.Name, "not checked yet"
	if snapshot, seen := states[game.UniverseID]; seen {
		updated = snapshot.Updated
		if name == "" {
			name = snapshot.Name
		}
	}
	if name == "" {
		name = "(name loads on first check)"
	}
	return fmt.Sprintf("[%s] %-12d %s  (updated %s)", status, game.PlaceID, name, updated)
}
