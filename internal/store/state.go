package store

import "github.com/yourpovv/AstralRbxTracker/internal/tracker"

type StateFile struct {
	Path string
}

var _ tracker.StateStore = StateFile{}

func (f StateFile) Load() (map[int64]tracker.Snapshot, error) {
	states := map[int64]tracker.Snapshot{}
	if err := readJSON(f.Path, &states); err != nil {
		return nil, err
	}
	return states, nil
}

func (f StateFile) Save(states map[int64]tracker.Snapshot) error {
	return writeJSON(f.Path, states)
}
