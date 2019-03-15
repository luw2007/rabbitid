package endpoints

import "context"

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"

	"github.com/luw2007/rabbitid/cmd/idHttp/service"
)

type NextRequest struct {
	APP string
	DB  string
}

type NextResponse struct {
	Id  int64  `json:"id,omitempty"`
	Msg string `json:"msg,omitempty"`
}

func MakeNextEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NextRequest)
		id, msg := s.Next(ctx, req.APP, req.DB)
		return NextResponse{Id: id, Msg: msg}, nil
	}
}

type LastRequest struct {
	Name string
}
type LastResponse struct {
	Id  int64  `json:"id,omitempty"`
	Msg string `json:"msg,omitempty"`
}

func MakeLastEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(LastRequest)
		id, msg := s.Last(ctx, req.Name)
		return LastResponse{Id: id, Msg: msg}, nil
	}
}

type RemainderRequest struct {
	Name string
}
type RemainderResponse struct {
	Id  int64  `json:"id,omitempty"`
	Msg string `json:"msg,omitempty"`
}

func MakeRemainderEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(RemainderRequest)
		id, msg := s.Remainder(ctx, req.Name)
		return RemainderResponse{Id: id, Msg: msg}, nil
	}
}

type MaxRequest struct {
	Name string
}
type MaxResponse struct {
	Id  int64  `json:"id,omitempty"`
	Msg string `json:"msg,omitempty"`
}

func MakeMaxEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(MaxRequest)
		id, msg := s.Max(ctx, req.Name)
		return MaxResponse{Id: id, Msg: msg}, nil
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
