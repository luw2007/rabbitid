package http

import (
	"context"
	"encoding/json"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/luw2007/rabbitid/cmd/idHttp/endpoints"
)

func NewHTTPHandler(endpoints endpoints.Endpoints) http.Handler {
	m := http.NewServeMux()
	m.Handle("/next", httptransport.NewServer(endpoints.Next, DecodeNextRequest, EncodeNextResponse))
	m.Handle("/last", httptransport.NewServer(endpoints.Last, DecodeLastRequest, EncodeLastResponse))
	m.Handle("/remainder", httptransport.NewServer(endpoints.Remainder, DecodeRemainderRequest, EncodeRemainderResponse))
	m.Handle("/max", httptransport.NewServer(endpoints.Max, DecodeMaxRequest, EncodeMaxResponse))
	return m
}
func DecodeNextRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.Request
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
func EncodeNextResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
func DecodeLastRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.Request
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
func EncodeLastResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
func DecodeRemainderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.Request
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
func EncodeRemainderResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
func DecodeMaxRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.Request
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
func EncodeMaxResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
