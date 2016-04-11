package main

import (
	// "database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
)

func main() {
	dbUrl := os.Getenv("DATABASE_URL") + `?sslmode=disable`

	log.WithFields(log.Fields{
		"url": dbUrl,
	}).Info("db URL")
	channel := GetEvents(dbUrl, 0)
	for {
		event := <-channel
		log.WithFields(log.Fields{
			"event": event,
		}).Info("received event")

	}
}

type Event struct {
	ID          int       `json:"id"`
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	EventOffset int       `json:"event_offset"`
	Timestamp   time.Time `json:"timestamp"`
	Payload     struct {
		Value int `json:"value"`
	} `json:"payload"`
}

type EventRow struct {
	ID          int    `db:"id"`
	EventID     string `db:"event_id"`
	EventType   string `db:"event_type"`
	EventOffset int    `db:"event_offset"`
	Timestamp   time.Time
	Payload     types.JSONText
}

type JSONEvent struct {
	ID          int    `json:"id"`
	EventID     string `json:"event_id"`
	EventType   string `json:"event_type"`
	EventOffset int    `json:"event_offset"`
	Timestamp   string `json:"timestamp"`
	Payload     struct {
		Value int `json:"value"`
	} `json:"payload"`
}

func (e EventRow) ToEvent() (Event, error) {
	var payload struct {
		Value int `json:"value"`
	}

	err := json.Unmarshal([]byte(e.Payload), &payload)
	if err != nil {
		return Event{}, err
	}

	return Event{
		ID:          e.ID,
		EventID:     e.EventID,
		EventType:   e.EventType,
		EventOffset: e.EventOffset,
		Timestamp:   e.Timestamp,
		Payload:     payload,
	}, nil

}

func (e *Event) UnmarshalJSON(data []byte) error {
	var eventJson JSONEvent

	err := json.Unmarshal(data, &eventJson)
	if err != nil {
		return err
	}

	// loc, err := time.LoadLocation("UTC")
	// if err != nil {
	// 	return fmt.Errorf("Error creating location %v", err)
	// }

	timestamp, err := time.Parse("2006-01-02T15:04:05", eventJson.Timestamp)
	// timestamp, err := pq.ParseTimestamp(loc, eventJson.Timestamp)
	if err != nil {
		return fmt.Errorf("Error parsing pq timestamp %s: %v", eventJson.Timestamp, err)
	}

	e.ID = eventJson.ID
	e.EventID = eventJson.EventID
	e.EventType = eventJson.EventType
	e.EventOffset = eventJson.EventOffset
	e.Timestamp = timestamp
	e.Payload = eventJson.Payload

	return nil
}

func GetEvents(dbUrl string, lastSeenID int) <-chan Event {
	channel := make(chan Event)

	go func() {
		lastSeenID := lastSeenID
		catchUp := func() {
			db, err := sqlx.Connect("postgres", dbUrl)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Fatal("sqlx.Connect")
			}

			eventRow := EventRow{}
			rows, err := db.Queryx("SELECT * FROM events WHERE id > $1", lastSeenID)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Fatal("db.Queryx")
			}
			for rows.Next() {
				err := rows.StructScan(&eventRow)
				if err != nil {
					log.WithFields(log.Fields{
						"err": err,
					}).Fatal("rows.StructScan")
				}

				event, err := eventRow.ToEvent()
				if err != nil {
					log.WithFields(log.Fields{
						"err":      err,
						"eventRow": eventRow,
					}).Fatal("eventRow.ToEvent")
				}

				channel <- event
				lastSeenID = event.ID
			}
		}
		catchUp()

		var (
			shouldCatchUp    bool
			shouldCatchUpMut sync.RWMutex
		)

		// Start listening
		listener := pq.NewListener(dbUrl, 5*time.Second, 1*time.Minute, func(listenerEvent pq.ListenerEventType, err error) {
			switch listenerEvent {
			case pq.ListenerEventReconnected:
				shouldCatchUpMut.Lock()
				shouldCatchUp = true
				shouldCatchUpMut.Unlock()
			}
			log.WithFields(log.Fields{
				"event": listenerEvent,
			}).Info("event callback")
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("event callback")
			}
		})
		defer listener.Close()

		err := listener.Listen("event_stream")
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal("listener listen")
		}

		listenerChan := listener.NotificationChannel()

		for {
			shouldCatchUpMut.RLock()
			if shouldCatchUp {
				catchUp()
			}
			shouldCatchUpMut.RUnlock()

			eventNotification := <-listenerChan

			event := Event{}

			err := json.Unmarshal([]byte(eventNotification.Extra), &event)
			if err != nil {
				log.WithFields(log.Fields{
					"err":          err,
					"notification": eventNotification,
				}).Fatal("json.Unmarshal event")
			}
			channel <- event
		}
	}()

	return channel
}
