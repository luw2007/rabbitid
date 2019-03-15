package main

import (
	"log"
	"sync"

	"github.com/luw2007/rabbitid/cmd/idHttp/conf"
	"github.com/luw2007/rabbitid/cmd/idRedis/handle"
	"github.com/tidwall/redcon"
)

var (
	server *redcon.Server
)

func main() {
	config := conf.Init()

	handler := handle.NewRedisHandler(config)
	server = redcon.NewServer(config.Server.Address, handler.Serve, handler.Connected, handler.Closed)

	var failed bool
	var wg sync.WaitGroup
	wg.Add(1)
	ch := make(chan error)
	go func(srv *redcon.Server) {
		defer wg.Done()
		if err := srv.ListenServeAndSignal(ch); err != nil {
			log.Fatal(err)
		}
	}(server)
	if err := <-ch; err != nil {
		log.Printf("error: %s", err.Error())
		failed = true
		return
	}
	log.Printf("server listening at %s", config.Server.Address)

	if failed {
		log.Println("Failed to start")
		handler.ShutDown()
		server.Close()
	}
	wg.Wait()

	log.Println("Graceful shutdown")
}
