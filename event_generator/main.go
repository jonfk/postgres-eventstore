package main

import (
	//"database/sql"
	"encoding/json"
	// "fmt"
	"math/rand"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
)

func main() {
	app := cli.NewApp()
	app.Name = "event_generator"
	app.Usage = "generates events and saves to postgres"
	app.Action = action

	app.Run(os.Args)
}

type Payload struct {
	Value int `json:"value"`
}

type Event struct {
	ID          int    `db:"omitempty"`
	EventID     string `db:"event_id"`
	EventType   string `db:"event_type"`
	EventOffset int    `db:"event_offset"`
	Timestamp   time.Time
	Payload     types.JSONText
}

func action(c *cli.Context) {
	dbUrl := os.Getenv("DATABASE_URL") + `?sslmode=disable`
	db, err := sqlx.Connect("postgres", dbUrl)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Failed to connect to db")
	}

	tx := db.MustBegin()
	for i := 0; i < 100; i++ {
		// tx.MustExec("INSERT INTO events (event_id, event_type, event_offset, timestamp, payload) VALUES ($1, $2, $3, $4, $5)", eventId, eventType, i, timestamp, JSONPayload)
		event := generateEvent(i)
		_, err := tx.NamedExec("INSERT INTO events (event_id, event_type, event_offset, timestamp, payload) VALUES (:event_id, :event_type, :event_offset, :timestamp, :payload)", &event)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal("named exec")
		}
	}
	tx.Commit()

}

func generateEvent(offset int) Event {
	eventId := uuid.NewV4()
	eventType := "positive"
	isPositiveEvent := rand.Int()%2 == 0
	if !isPositiveEvent {
		eventType = "negative"
	}
	timestamp := time.Now()
	value := rand.Intn(100)
	if !isPositiveEvent {
		value = -value
	}
	payload := Payload{Value: value}
	JSONPayload, err := json.Marshal(payload)

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("json marshal error in generate event")
	}

	return Event{
		EventID:     eventId.String(),
		EventType:   eventType,
		EventOffset: offset,
		Timestamp:   timestamp,
		Payload:     JSONPayload,
	}
}
