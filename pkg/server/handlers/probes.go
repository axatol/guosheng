package handlers

import (
	"context"
	"net/http"

	"github.com/axatol/go-utils/httputil"
)

type ReadyChecker interface {
	Ready(ctx context.Context) bool
}

func Ping(rc ReadyChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := httputil.Response[any]{}
		defer res.Write(w)

		if !rc.Ready(r.Context()) {
			res.SetStatus(http.StatusServiceUnavailable)
			res.SetMessage("service not ready, please try again later")
		} else {
			res.SetStatus(http.StatusOK)
		}
	}
}

type HealthChecker interface {
	Health(ctx context.Context) (any, error)
}

func Health(hc HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := httputil.Response[any]{}
		defer res.Write(w)

		metadata, err := hc.Health(r.Context())
		res.SetData(&metadata)
		if err != nil {
			res.SetStatus(http.StatusServiceUnavailable)
			res.SetError(err)
		}
	}
}
