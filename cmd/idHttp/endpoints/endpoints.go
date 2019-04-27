package endpoints

import "context"

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"

	"github.com/luw2007/rabbitid/cmd/idHttp/service"
)

type Request struct {
	APP string `json:"app"`
	DB  string `json:"db"`
}

type Response struct {
	Id  int64  `json:"id,omitempty"`
	Msg string `json:"msg,omitempty"`
}

func MakeNextEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(Request)
		id, msg := s.Next(ctx, req.APP, req.DB)
		return Response{Id: id, Msg: msg}, nil
	}
}

func MakeLastEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(Request)
		id, msg := s.Last(ctx, req.APP, req.DB)
		return Response{Id: id, Msg: msg}, nil
	}
}

func MakeRemainderEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(Request)
		id, msg := s.Remainder(ctx, req.APP, req.DB)
		return Response{Id: id, Msg: msg}, nil
	}
}

func MakeMaxEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(Request)
		id, msg := s.Max(ctx, req.APP, req.DB)
		return Response{Id: id, Msg: msg}, nil
	}
}

type Endpoints struct {
	Next      endpoint.Endpoint
	Last      endpoint.Endpoint
	Remainder endpoint.Endpoint
	Max       endpoint.Endpoint
}

func New(svc service.Service, logger log.Logger) Endpoints {
	var nextEndpoint endpoint.Endpoint
	nextEndpoint = MakeNextEndpoint(svc)
	nextEndpoint = LoggingMiddleware(log.With(logger, "method", "Next"))(nextEndpoint)

	var lastEndpoint endpoint.Endpoint
	lastEndpoint = MakeLastEndpoint(svc)
	lastEndpoint = LoggingMiddleware(log.With(logger, "method", "Last"))(lastEndpoint)

	var remainderEndpoint endpoint.Endpoint
	remainderEndpoint = MakeRemainderEndpoint(svc)
	remainderEndpoint = LoggingMiddleware(log.With(logger, "method", "Remainder"))(remainderEndpoint)

	var maxEndpoint endpoint.Endpoint
	maxEndpoint = MakeMaxEndpoint(svc)
	maxEndpoint = LoggingMiddleware(log.With(logger, "method", "Max"))(maxEndpoint)

	return Endpoints{
		Next:      nextEndpoint,
		Last:      lastEndpoint,
		Remainder: remainderEndpoint,
		Max:       maxEndpoint,
	}
}
