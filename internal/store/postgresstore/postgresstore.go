package postgresstore

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/catalystcommunity/app-utils-go/env"
	"github.com/catalystcommunity/app-utils-go/logging"
	_ "github.com/catalystcommunity/service-go-hello-api/internal/store"
	. "github.com/catalystcommunity/service-go-hello-api/internal/store/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	goose "github.com/pressly/goose/v3"
	sqlhooks "github.com/qustavo/sqlhooks/v2"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	// "github.com/catalystcommunity/service-go-hello-api/cmd/config"
)

// global db
var bunDb *bun.DB

// var DatabaseUri = env.GetEnvOrDefault("DATABASE_URI", "host=service-go-connect-api-postgresql.skaffold-service-go-connect-api.svc port=5432 user=devdbuser password=devdbpass dbname=devdb sslmode=disable")
var DatabaseUri = env.GetEnvOrDefault("DATABASE_URI", "postgres://devdbuser:devdbpass@service-go-connect-api-postgresql.skaffold-service-go-connect-api.svc:5432/devdb?sslmode=disable")

// sql files embedded at compile time, used by goose
//
//go:embed migrations/*.sql
var embedMigrations embed.FS

type PostgresStore struct{}

func (s PostgresStore) Initialize() (func(), error) {
	var err error
	logging.Log.WithFields(logrus.Fields{"path": "postgresstore.go"}).Info("URI: ", DatabaseUri)
	sql.Register("apphooks", sqlhooks.Wrap(pq.Driver{}, &Hooks{}))
	pgdb, err := sql.Open("postgres", DatabaseUri)
	if err != nil {
		return nil, err
	}

	// Create a Bun db on top of it.
	bunDb = bun.NewDB(pgdb, pgdialect.New())

	// Print all queries to stdout.
	// bunDb.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	// set goose file system to use the embedded migrations
	goose.SetBaseFS(embedMigrations)
	err = goose.Up(pgdb, "migrations")
	if err != nil {
		return nil, err
	}

	return func() {
		println(fmt.Sprintf("Closing DB!"))
		pgdb.Close()
	}, nil

}

func (s PostgresStore) Hello(data HelloId) (*Hello, *ApiError) {
	ctx := context.Background()

	helloResponse := new(Hello)
	queryErr := bunDb.NewSelect().Model(helloResponse).
		Where("id = ?", data.Id).Scan(ctx)

	if queryErr == sql.ErrNoRows {
		return nil, &ApiError{404, "id was not found"}
	}
	if queryErr != nil && queryErr != sql.ErrNoRows {
		return nil, &ApiError{500, "could not select id"}
	}

	return helloResponse, nil
}

func (s PostgresStore) CreateHello(data NewHello) (*Hello, *ApiError) {
	ctx := context.Background()

	helloResponse := new(Hello)
	count, queryErr := bunDb.NewSelect().Model(helloResponse).
		Where("name = ?", data.Name).ScanAndCount(ctx)

	println(fmt.Sprintf("In CreateHello: %s, count: %d", queryErr, count))
	if queryErr != nil && queryErr != sql.ErrNoRows {
		return nil, &ApiError{500, "could not query on that name"}
	}
	if count == 0 {
		newUuid, _ := uuid.NewRandom()
		helloResponse.Id = newUuid.String()
		helloResponse.Name = data.Name
		_, err := bunDb.NewInsert().Model(helloResponse).Exec(ctx)
		if err != nil {
			return nil, &ApiError{500, err.Error()}
		}
	}

	return helloResponse, nil
}
