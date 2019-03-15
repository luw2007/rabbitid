package store

import (
	"context"
	"log"
	"os"

	kitlog "github.com/go-kit/kit/log"
)

func ExampleNewEtcd() {
	logger := kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stderr))
	db := NewEtcd(testURI, logger)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err := db.Ping(ctx)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleNewRedis() {
	logger := kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stderr))
	db := NewRedis(testRedis, logger)
	err := db.Ping(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}
