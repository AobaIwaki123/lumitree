package timetree

import (
	"time"

	"github.com/AobaIwaki123/lumi-spec/event"
	"github.com/AobaIwaki123/lumitree/pkg/model"
)

// ToEventFeed converts the normalized TimeTree response to the shared
// lumi-spec event-feed contract. The HTTP API can continue to expose the
// source-native response while callers that need interoperability use this
// adapter directly.
func ToEventFeed(calendarID string, response *model.EventListResponse) event.Feed {
	feed := event.Feed{
		SchemaVersion: event.FeedSchemaVersion,
		Source:        "lumitree",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		Events:        make([]event.Candidate, 0),
	}
	if response == nil {
		return feed
	}
	feed.Events = make([]event.Candidate, 0, len(response.Events))
	for _, item := range response.Events {
		candidate := model.ToEventCandidate(calendarID, item)
		if candidate.ID != "lumitree:" {
			feed.Events = append(feed.Events, candidate)
		}
	}
	return feed
}
