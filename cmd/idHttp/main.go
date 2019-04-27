package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/luw2007/rabbitid/cmd/idHttp/conf"
	"github.com/luw2007/rabbitid/cmd/idHttp/service"
	"github.com/luw2007/rabbitid/store"
)

type Request struct {
	APP string `json:"app"`
	DB  string `json:"db"`
}

type Response struct {
	Code int64  `json:"code,omitempty"`
	ID   int64  `json:"id"`
	Msg  string `json:"msg,omitempty"`
}

func main() {
	g := gin.Default()
	config := conf.Init()
	logger := config.Logger.WithField("svc", "idhttp")

	db := store.NewStore(config.Store.Type, config.Store.URI, config.Generate.DataCenter, logger)
	svc := service.New(logger, db, config.Generate.Step, config.Generate.DataCenter, config.Store.Min, config.Store.Max)

	g.POST("/next", func(c *gin.Context) {
		var req Request
		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(200, Response{Code: -1, Msg: "argument error"})
			return
		}
		id, msg := svc.Next(context.Background(), req.APP, req.DB)
		c.JSON(200, Response{ID: id, Msg: msg})
	})
	g.POST("/last", func(c *gin.Context) {
		var req Request
		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(200, Response{Code: -1, Msg: "argument error"})
			return
		}
		id, msg := svc.Last(context.Background(), req.APP, req.DB)
		c.JSON(200, Response{ID: id, Msg: msg})
	})
	g.POST("/max", func(c *gin.Context) {
		var req Request
		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(200, Response{Code: -1, Msg: "argument error"})
			return
		}
		id, msg := svc.Max(context.Background(), req.APP, req.DB)
		c.JSON(200, Response{ID: id, Msg: msg})
	})
	g.POST("/remainder", func(c *gin.Context) {
		var req Request
		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(200, Response{Code: -1, Msg: "argument error"})
			return
		}
		id, msg := svc.Remainder(context.Background(), req.APP, req.DB)
		c.JSON(200, Response{ID: id, Msg: msg})
	})

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Info("transport", "HTTP", "addr", config.Server.Address)
		errs <- http.ListenAndServe(config.Server.Address, g)
	}()

	logger.Info("exit", <-errs)
}
