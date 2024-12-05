package utils

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
)

// ListenAndServe lets the server be closed whenever the context is closed.
func ListenAndServe(ctx context.Context, server *http.Server, startMsg string) error {
	go func() {
		<-ctx.Done()
		server.Close()
	}()
	logrus.Info(startMsg)
	return server.ListenAndServe()
}
