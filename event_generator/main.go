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

type PositivePayload struct {
	Value int `json:"value"`
}

type Event struct {
	ID          int
	EventID     string `db:"event_id"`
	EventType   string `db:"event_type"`
	EventOffset int    `db:"event_offset"`
	Timestamp   time.Time
	Payload     types.JSONText
}

func action(c *cli.Context) {
	// dbUser := os.Getenv("DATABASE_USER")
	// dbPass := os.Getenv("DATABASE_PASSWORD")
	// dbName := os.Getenv("DATABASE_NAME")
	dbUrl := os.Getenv("DATABASE_URL") + `?sslmode=disable`
	db, err := sqlx.Connect("postgres", dbUrl) //fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPass, dbName))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Failed to connect to db")
	}

	tx := db.MustBegin()
	for i := 0; i < 10; i++ {
		eventId := uuid.NewV4()
		eventType := "positive"
		timestamp := time.Now()
		payload := PositivePayload{Value: rand.Int()}
		JSONPayload, _ := json.Marshal(payload)
		tx.MustExec("INSERT INTO events (event_id, event_type, event_offset, timestamp, payload) VALUES ($1, $2, $3, $4, $5)", eventId, eventType, i, timestamp, JSONPayload)
	}
	tx.Commit()

}
