package store

import (
	"context"

	"github.com/sirupsen/logrus"
)

func ExampleNewEtcd() {
	log := logrus.NewEntry(logrus.New())
	db := NewEtcd(testURI, log)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err := db.Ping(ctx)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleNewRedis() {
	log := logrus.NewEntry(logrus.New())
	db := NewRedis(testRedis, log)
	err := db.Ping(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}
