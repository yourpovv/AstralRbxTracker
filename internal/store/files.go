// game list and per-game state as json files
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func readJSON(path string, out any) error {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("can't read %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("can't read %s: %w", path, err)
	}
	return nil
}

// write then swap
func writeJSON(path string, payload any) error {
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("can't write to %s: %w", path, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("can't write to %s: %w", path, err)
	}
	return nil
}
