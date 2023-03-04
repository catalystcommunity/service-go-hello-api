package internal

import (
	"encoding/json"
	"net/http"

	. "github.com/catalystsquad/service-go-hello-api/internal/store"
	. "github.com/catalystsquad/service-go-hello-api/internal/store/models"
	"github.com/uptrace/bunrouter"
)

func GenericHandler[M ApiModel, R ApiModel, S StoreHandler[M, R]](apiModel M, retModel R, storeHandler S) bunrouter.HandlerFunc {
	retFunc := func(w http.ResponseWriter, req bunrouter.Request) error {
		err := json.NewDecoder(req.Body).Decode(&apiModel)
		if err != nil {
			w.WriteHeader(400)
			return bunrouter.JSON(w, &ApiError{400, "bad json provided"})
		}
		retModel, apiError := storeHandler(apiModel)
		if apiError != nil {
			w.WriteHeader(apiError.Code)
			return bunrouter.JSON(w, apiError)
		}
		return bunrouter.JSON(w, retModel)
	}
	return retFunc
}

func RegisterRoutes(router *bunrouter.Router) *bunrouter.Router {
	helloHandler := GenericHandler(HelloId{}, Hello{}, AppStore.Hello)
	router.POST("/hello", helloHandler)
	router.POST("/newhello", GenericHandler(NewHello{}, Hello{}, AppStore.CreateHello))
	return router
}

//func decodeJsonToModel[M ApiModel](req bunrouter.Request) (*M, error) {
//	model := new(M)
//	err := json.NewDecoder(req.Body).Decode(model)
//	if err != nil {
//		return nil, errors.New("bad Hello json provided")
//	}
//	return model, nil
//}
//
//func resultToJson[M ApiModel](w http.ResponseWriter, model M, apiError *ApiError) error {
//	if apiError != nil {
//		http.Error(w, apiError.Error, apiError.Code)
//		return errors.New(apiError.Error)
//	}
//	return bunrouter.JSON(w, model)
//}

//func RegisterRoutes(router *bunrouter.Router) *bunrouter.Router {
//	router.POST("/hello", HelloHandler)
//	router.POST("/newhello", CreateHelloHandler)
//	return router
//}
//
//func HelloHandler(w http.ResponseWriter, req bunrouter.Request) error {
//	var helloId HelloId
//	println(fmt.Sprintf("In Hello Handler: %s", req.Body))
//	err := json.NewDecoder(req.Body).Decode(&helloId)
//	if err != nil {
//		w.WriteHeader(400)
//		return bunrouter.JSON(w, ApiError{400, "bad HelloId json provided"})
//	}
//	model, apiError := AppStore.Hello(helloId)
//	if apiError != nil {
//		w.WriteHeader(apiError.Code)
//		return bunrouter.JSON(w, apiError)
//	}
//	return bunrouter.JSON(w, model)
//}
//
//func CreateHelloHandler(w http.ResponseWriter, req bunrouter.Request) error {
//	var newhello NewHello
//	println(fmt.Sprintf("In CreateHello Handler: %s", req.Body))
//	err := json.NewDecoder(req.Body).Decode(&newhello)
//	if err != nil {
//		w.WriteHeader(400)
//		return bunrouter.JSON(w, ApiError{400, "bad NewHello json provided"})
//	}
//	println(fmt.Sprintf("AppStore: %s", AppStore))
//	println(fmt.Sprintf("Calling CreateHello: %s", newhello))
//	model, apiError := AppStore.CreateHello(newhello)
//	if apiError != nil {
//		println(fmt.Sprintf("ApiError: %v", apiError))
//		w.WriteHeader(apiError.Code)
//		return bunrouter.JSON(w, apiError)
//	}
//	println(fmt.Sprintf("Returned Hello: %s", model))
//	return bunrouter.JSON(w, model)
//}
