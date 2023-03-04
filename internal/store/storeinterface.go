package store

import (
	. "github.com/catalystsquad/service-go-hello-api/internal/store/models"
)

var AppStore Store

type Store interface {
	Initialize() (deferredFunc func(), err error)

	Hello(HelloId) (*Hello, *ApiError)
	CreateHello(NewHello) (*Hello, *ApiError)
}
