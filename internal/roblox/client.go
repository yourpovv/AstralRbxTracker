// this client talks to roblox's public apis
package roblox

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/yourpovv/AstralRbxTracker/internal/tracker"
)

const (
	batchSize  = 50
	iconSize   = "512x512"
	iconFormat = "Png"
	iconReady  = "Completed"
)

var ErrPlaceNotFound = errors.New("place doesn't exist on Roblox, check the place ID")

type Endpoints struct {
	Universes  string
	Games      string
	Thumbnails string
}

var PublicEndpoints = Endpoints{
	Universes:  "https://apis.roblox.com/universes/v1/places/%d/universe",
	Games:      "https://games.roblox.com/v1/games",
	Thumbnails: "https://thumbnails.roblox.com/v1/games/icons",
}

type Client struct {
	HTTP      *http.Client
	Endpoints Endpoints
	Retry     Retry
	Log       *slog.Logger
}

func (c *Client) UniverseID(ctx context.Context, placeID int64) (int64, error) {
	var response struct {
		UniverseID int64 `json:"universeId"`
	}
	target := fmt.Sprintf(c.Endpoints.Universes, placeID)
	if err := c.getJSON(ctx, target, &response); err != nil {
		return 0, err
	}
	if response.UniverseID == 0 {
		return 0, fmt.Errorf("place %d: %w", placeID, ErrPlaceNotFound)
	}
	return response.UniverseID, nil
}

var _ tracker.Source = (*Client)(nil)

// 50 games per request avoids roblox's rate limit
func (c *Client) Snapshots(
	ctx context.Context, universeIDs []int64,
) (map[int64]tracker.Snapshot, error) {
	snapshots, err := c.games(ctx, universeIDs)
	if err != nil {
		return nil, err
	}

	iconURLs, err := c.icons(ctx, universeIDs)
	if err != nil {
		// we won't block alerts for icon outages
		c.Log.Warn("game icons unavailable", "err", err)
		return snapshots, nil
	}
	for universeID, snapshot := range snapshots {
		snapshot.IconURL = iconURLs[universeID]
		snapshots[universeID] = snapshot
	}
	return snapshots, nil
}

func (c *Client) games(ctx context.Context, universeIDs []int64) (map[int64]tracker.Snapshot, error) {
	snapshots := map[int64]tracker.Snapshot{}
	for _, batch := range batches(universeIDs) {
		var response gamesResponse
		query := url.Values{"universeIds": {joined(batch)}}
		if err := c.getJSON(ctx, c.Endpoints.Games+"?"+query.Encode(), &response); err != nil {
			return nil, err
		}
		for _, entry := range response.Data {
			if entry.Updated == "" {
				return nil, fmt.Errorf("roblox game %d data is missing its update time", entry.ID)
			}
			snapshots[entry.ID] = entry.snapshot()
		}
	}
	return snapshots, nil
}

func (c *Client) icons(ctx context.Context, universeIDs []int64) (map[int64]string, error) {
	iconURLs := map[int64]string{}
	for _, batch := range batches(universeIDs) {
		var response iconsResponse
		query := url.Values{
			"universeIds": {joined(batch)},
			"size":        {iconSize},
			"format":      {iconFormat},
		}
		target := c.Endpoints.Thumbnails + "?" + query.Encode()
		if err := c.getJSON(ctx, target, &response); err != nil {
			return nil, err
		}
		for _, entry := range response.Data {
			if entry.State == iconReady {
				iconURLs[entry.TargetID] = entry.ImageURL
			}
		}
	}
	return iconURLs, nil
}

type gamesResponse struct {
	Data []gameEntry `json:"data"`
}

type gameEntry struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Updated     string `json:"updated"`
	Playing     int    `json:"playing"`
}

func (e gameEntry) snapshot() tracker.Snapshot {
	return tracker.Snapshot{
		Name:        e.Name,
		Description: e.Description,
		Updated:     e.Updated,
		Playing:     e.Playing,
	}
}

type iconsResponse struct {
	Data []struct {
		TargetID int64  `json:"targetId"`
		State    string `json:"state"`
		ImageURL string `json:"imageUrl"`
	} `json:"data"`
}

func batches(universeIDs []int64) [][]int64 {
	result := [][]int64{}
	for start := 0; start < len(universeIDs); start += batchSize {
		result = append(result, universeIDs[start:min(start+batchSize, len(universeIDs))])
	}
	return result
}

func joined(universeIDs []int64) string {
	parts := make([]string, len(universeIDs))
	for i, universeID := range universeIDs {
		parts[i] = strconv.FormatInt(universeID, 10)
	}
	return strings.Join(parts, ",")
}
