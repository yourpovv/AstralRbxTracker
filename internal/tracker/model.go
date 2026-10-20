package tracker

import "time"

// a game to watch (empty Name will use the published game name)
type Game struct {
	PlaceID    int64  `json:"place_id"`
	UniverseID int64  `json:"universe_id"`
	Name       string `json:"name,omitempty"`
	Enabled    bool   `json:"enabled"`
}

// what roblox shares about a game
type Snapshot struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Updated     string `json:"updated"`
	Playing     int    `json:"playing"`
	IconURL     string `json:"icon_url"`
}

func (s Snapshot) WithIconFrom(previous Snapshot) Snapshot {
	if s.IconURL == "" {
		s.IconURL = previous.IconURL
	}
	return s
}

// a confirmed public diff by roblox
type Change struct {
	Label   string
	Summary string
}

type Update struct {
	Game       Game
	Previous   Snapshot
	Current    Snapshot
	Changes    []Change
	DetectedAt time.Time
}

func (u Update) Published() bool {
	return u.Previous.Updated != u.Current.Updated
}

func (u Update) IsNews() bool {
	return u.Published() || len(u.Changes) > 0
}

func (u Update) DisplayName() string {
	if u.Game.Name != "" {
		return u.Game.Name
	}
	return u.Current.Name
}
