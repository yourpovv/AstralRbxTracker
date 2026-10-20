package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/yourpovv/AstralRbxTracker/internal/tracker"
)

const (
	updateColor  = 0x7C3AED
	detailsColor = 0x14B8A6
	footerText   = "Astral Tracker · Place %d"
	gameURL      = "https://www.roblox.com/games/%d"
	fieldLimit   = 1024

	noPublicChanges = "Nothing public, so probably scripts or the map"
)

type message struct {
	Username string  `json:"username"`
	Embeds   []embed `json:"embeds"`
}

type embed struct {
	Author      author     `json:"author"`
	Title       string     `json:"title"`
	URL         string     `json:"url"`
	Description string     `json:"description"`
	Color       int        `json:"color"`
	Fields      []field    `json:"fields"`
	Thumbnail   *thumbnail `json:"thumbnail,omitempty"`
	Footer      footer     `json:"footer"`
	Timestamp   string     `json:"timestamp"`
}

type author struct {
	Name string `json:"name"`
}

type field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type thumbnail struct {
	URL string `json:"url"`
}

type footer struct {
	Text string `json:"text"`
}

func newEmbed(update tracker.Update) embed {
	result := headline(update)
	result.Title = update.DisplayName()
	result.URL = fmt.Sprintf(gameURL, update.Game.PlaceID)
	result.Fields = fields(update)
	result.Footer = footer{Text: fmt.Sprintf(footerText, update.Game.PlaceID)}
	result.Timestamp = update.DetectedAt.UTC().Format(time.RFC3339)
	if update.Current.IconURL != "" {
		result.Thumbnail = &thumbnail{URL: update.Current.IconURL}
	}
	return result
}

func headline(update tracker.Update) embed {
	if update.Published() {
		return embed{
			Author:      author{Name: "🛰️ New version live"},
			Description: "Published " + discordTime(update.Current.Updated),
			Color:       updateColor,
		}
	}
	return embed{
		Author:      author{Name: "✏️ Details edited"},
		Description: "The game page changed, but no new version was published.",
		Color:       detailsColor,
	}
}

func fields(update tracker.Update) []field {
	result := []field{{Name: "Players", Value: withThousands(update.Current.Playing), Inline: true}}
	if update.Published() {
		result = append(result, field{Name: "Previous version", Value: discordTime(update.Previous.Updated), Inline: true})
	}
	return append(result, field{Name: "Changes", Value: whatChanged(update.Changes)})
}

// shows in each reader's own time zone, raw text when it isn't a timestamp
func discordTime(timestamp string) string {
	parsed, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "`" + timestamp + "`"
	}
	return fmt.Sprintf("<t:%d:R>", parsed.Unix())
}

func whatChanged(changes []tracker.Change) string {
	if len(changes) == 0 {
		return noPublicChanges
	}
	summaries := make([]string, len(changes))
	for i, change := range changes {
		summaries[i] = change.Summary
	}
	return truncate(strings.Join(summaries, "\n"), fieldLimit)
}

func truncate(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit])
}

func withThousands(number int) string {
	digits := fmt.Sprint(number)
	var grouped strings.Builder
	for i, digit := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			grouped.WriteByte(',')
		}
		grouped.WriteRune(digit)
	}
	return grouped.String()
}
