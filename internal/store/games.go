package store

import (
	"encoding/json"
	"fmt"

	"github.com/yourpovv/AstralRbxTracker/internal/tracker"
)

type GameFile struct {
	Path string
}

var _ tracker.GameList = GameFile{}

type gameFileContent struct {
	Games []json.RawMessage `json:"games"`
}

func (f GameFile) Games() ([]tracker.Game, error) {
	var content gameFileContent
	if err := readJSON(f.Path, &content); err != nil {
		return nil, err
	}

	games := make([]tracker.Game, 0, len(content.Games))
	for _, raw := range content.Games {
		game := tracker.Game{Enabled: true}
		if err := json.Unmarshal(raw, &game); err != nil {
			return nil, fmt.Errorf("can't read %s: %w", f.Path, err)
		}
		if game.PlaceID == 0 || game.UniverseID == 0 {
			return nil, fmt.Errorf("%s: every game needs place_id and universe_id", f.Path)
		}
		games = append(games, game)
	}
	return games, nil
}

func (f GameFile) Save(games []tracker.Game) error {
	return writeJSON(f.Path, struct {
		Games []tracker.Game `json:"games"`
	}{games})
}
