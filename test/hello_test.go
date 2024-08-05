package test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/catalystcommunity/app-utils-go/env"
	"github.com/catalystcommunity/app-utils-go/logging"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/catalystcommunity/service-go-hello-api/internal"
	. "github.com/catalystcommunity/service-go-hello-api/internal/store"
	. "github.com/catalystcommunity/service-go-hello-api/internal/store/models"
	"github.com/catalystcommunity/service-go-hello-api/internal/store/postgresstore"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bunrouter"
)

func jsonByteBuffer(s string) *bytes.Buffer {
	return bytes.NewBuffer([]byte(s))
}
func executeRequest(router *bunrouter.Router, req *http.Request) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	return res
}

type HelloSuite struct {
	suite.Suite
	deferredFunc func()
	bunDb        *bun.DB
}

func (s *HelloSuite) SetupSuite() {
	DatabaseUri := env.GetEnvOrDefault("DATABASE_URI", "postgres://devdbuser:devdbpass@localhost:5432/devdb?sslmode=disable")
	AppStore = postgresstore.PostgresStore{}
	deferredFunc, err := AppStore.Initialize()
	if err != nil {
		logging.Log.WithField("error", fmt.Sprintf("%s", err)).Fatal("Could not initialize store interface")
		panic(err)
	}
	if deferredFunc != nil {
		s.deferredFunc = deferredFunc
	}
	pgdb, err := sql.Open("postgres", DatabaseUri)
	if err != nil {
		noDb := "can't open DB"
		println(noDb)
		panic(errors.New(noDb))
		return
	}

	// Create a Bun db on top of it.
	s.bunDb = bun.NewDB(pgdb, pgdialect.New())
}

func (s *HelloSuite) TearDownSuite() {
	s.deferredFunc()
}

func (s *HelloSuite) SetupTest() {
}

func TestHelloSuite(t *testing.T) {
	suite.Run(t, new(HelloSuite))
}

func deferredDeleteHello(s *HelloSuite, id string) {
	_, _ = s.bunDb.NewDelete().
		Table("hellos").
		Where("id = ?", id).
		Exec(context.Background())
	return
}

func (s *HelloSuite) TestHelloNotFound() {
	newUuid, _ := uuid.NewRandom()
	router := bunrouter.New()
	router.POST("/hello", GenericHandler(HelloId{}, Hello{}, AppStore.Hello))
	req, _ := http.NewRequest("POST", "/hello", jsonByteBuffer(fmt.Sprintf(`{"id": "%s"}`, newUuid.String())))
	response := executeRequest(router, req)

	expected := ApiError{Code: 404, Error: "id was not found"}
	var apiError ApiError
	_ = json.NewDecoder(bytes.NewBuffer(response.Body.Bytes())).Decode(&apiError)

	require.Equal(s.T(), expected, apiError, fmt.Sprintf("id should not be found, but error does not say this. Instead: %s", response.Body.Bytes()))
}

func (s *HelloSuite) TestHelloCreateAndGet() {
	subject := "DeleteThisDevTestPerson"
	router := bunrouter.New()
	router.POST("/newhello", GenericHandler(NewHello{}, Hello{}, AppStore.CreateHello))
	req := httptest.NewRequest("POST", "/newhello", jsonByteBuffer(fmt.Sprintf(`{"name": "%s"}`, subject)))
	response := executeRequest(router, req)

	var hello Hello
	_ = json.NewDecoder(bytes.NewBuffer(response.Body.Bytes())).Decode(&hello)
	var apiError ApiError
	_ = json.NewDecoder(bytes.NewBuffer(response.Body.Bytes())).Decode(&apiError)
	require.Empty(s.T(), apiError.Error, fmt.Sprintf("error should be empty. error: \n%v", apiError.Error))
	// Clean up this created item
	defer deferredDeleteHello(s, hello.Id)
	expectedId := hello.Id
	require.Equal(s.T(), subject, hello.Name, fmt.Sprintf("returned name from CreateHello is incorrect.\nExpected: %s, Returned: %s", subject, hello.Name))

	router.POST("/hello", GenericHandler(HelloId{}, Hello{}, AppStore.Hello))
	req, _ = http.NewRequest("POST", "/hello", jsonByteBuffer(fmt.Sprintf(`{"id": "%s"}`, hello.Id)))
	response = executeRequest(router, req)

	hello = Hello{}
	_ = json.NewDecoder(bytes.NewBuffer(response.Body.Bytes())).Decode(&hello)
	_ = json.NewDecoder(bytes.NewBuffer(response.Body.Bytes())).Decode(&apiError)
	println(fmt.Sprintf("Returned Hello: %s", hello))

	require.Empty(s.T(), apiError.Error, fmt.Sprintf("error should be empty. error: \n%v", apiError.Error))
	require.NotEmpty(s.T(), hello.Id, "returned ID from Hello is empty")
	require.Equal(s.T(), expectedId, hello.Id, fmt.Sprintf("returned id from Hello is incorrect.\nExpected: %s, Returned: %s", subject, hello.Id))
	require.Equal(s.T(), subject, hello.Name, fmt.Sprintf("returned name from Hello is incorrect.\nExpected: %s, Returned: %s", subject, hello.Name))
}
