package main

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/lib/pq"
)

func main() {
	dbUrl := os.Getenv("DATABASE_URL") + `?sslmode=disable`

	log.WithFields(log.Fields{
		"url": dbUrl,
	}).Info("db URL")
	listener := pq.NewListener(dbUrl, 5*time.Second, 1*time.Minute, func(event pq.ListenerEventType, err error) {
		log.WithFields(log.Fields{
			"event": event,
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

	channel := listener.NotificationChannel()

	for {
		event := <-channel
		log.WithFields(log.Fields{
			"event": event,
		}).Info("event received")
	}
}
