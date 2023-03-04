package models

type ApiError struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type ApiModel interface {
	Hello |
		HelloId |
		NewHello
}

type StoreHandler[M ApiModel, R ApiModel] func(M) (*R, *ApiError)

type Hello struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type HelloId struct {
	Id string `json:"id"`
}

type NewHello struct {
	Name string `json:"name"`
}
