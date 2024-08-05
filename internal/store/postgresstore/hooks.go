package postgresstore

import (
	"context"
	"fmt"
	"github.com/catalystcommunity/app-utils-go/logging"
	"github.com/sirupsen/logrus"
	"time"
)

// Hooks satisfies the sqlhook.Hooks interface
type Hooks struct{}

// Before hook will print the query with it's args and return the context with the timestamp
func (h *Hooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	logging.Log.WithFields(logrus.Fields{"query": query, "args": args}).Info("query log")
	return context.WithValue(ctx, "begin", time.Now()), nil
}

// After hook will get the timestamp registered on the Before hook and print the elapsed time
func (h *Hooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	begin := ctx.Value("begin").(time.Time)
	fmt.Printf(". took: %s\n", time.Since(begin))
	return ctx, nil
}
