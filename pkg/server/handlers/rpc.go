package handlers

import (
	"context"
	"fmt"
	"net/http"
)

func Shutdown(cancel context.CancelCauseFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel(fmt.Errorf("received shutdown rpc"))
	}
}
