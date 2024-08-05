package cmd

import (
	"fmt"

	"net/http"

	"github.com/catalystcommunity/app-utils-go/errorutils"
	"github.com/catalystcommunity/app-utils-go/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"

	"github.com/catalystcommunity/service-go-hello-api/internal"
	. "github.com/catalystcommunity/service-go-hello-api/internal/store"
	"github.com/catalystcommunity/service-go-hello-api/internal/store/postgresstore"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs an service-go-hello-api service",
	Long:  `Runs an service-go-hello-api service with an service-go-hello-api health check.`,
	Run: func(cmd *cobra.Command, args []string) {
		logging.Init()
		config := initRunCmdConfig()

		// set and init store
		AppStore = postgresstore.PostgresStore{}
		deferredFunc, err := AppStore.Initialize()
		if err != nil {
			logging.Log.WithField("error", fmt.Sprintf("%s", err)).Fatal("Could not initialize store interface")
			panic(err)
		}
		// if the store has a deferred call, defer it
		if deferredFunc != nil {
			defer deferredFunc()
		}

		maybeStartHealthServer(config)
		startServer(config)
	},
}

type runCmdConfig struct {
	Port              int
	EnableHealthCheck bool
	HealthCheckPath   string
	HealthCheckPort   int
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().Int("port", 8080, "port for service-go-hello-api server")
	runCmd.PersistentFlags().Bool("enable_health_check", true, "when true, runs an http server on port 7000 that can be used for a health check for things like kubernetes with GET /health")
	runCmd.PersistentFlags().String("health_check_path", "/health", "path to serve health check on when health check is enabled")
	runCmd.PersistentFlags().Int("health_check_port", 7000, "port to serve health check on when health check is enabled")

	// bind flags
	err := viper.BindPFlags(runCmd.PersistentFlags())
	// die on error
	if err != nil {
		errorutils.PanicOnErr(nil, "error initializing configuration", err)
	}
}

func initRunCmdConfig() *runCmdConfig {
	// instantiate config struct
	config := &runCmdConfig{}

	config.Port = viper.GetInt("port")
	config.EnableHealthCheck = viper.GetBool("enable_health_check")
	config.HealthCheckPath = viper.GetString("health_check_path")
	config.HealthCheckPort = viper.GetInt("health_check_port")

	logging.Log.WithField("settings", fmt.Sprintf("%+v", *config)).Debug("viper settings")

	return config
}

func notFoundHandler(w http.ResponseWriter, req bunrouter.Request) error {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(
		w,
		"{code: 404, error:\"BunRouter can't find a route that matches %s\"}",
		req.URL.Path,
	)
	return nil
}

func methodNotAllowedHandler(w http.ResponseWriter, req bunrouter.Request) error {
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintf(
		w,
		"{code: 404, error:\"BunRouter does have a route that matches %s, "+
			"but it does not handle method %s\"}",
		req.URL.Path, req.Method,
	)
	return nil
}

func debugHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return bunrouter.JSON(w, bunrouter.H{
		"route":  req.Route(),
		"params": req.Params().Map(),
	})
}

func maybeStartHealthServer(config *runCmdConfig) {
	if config.EnableHealthCheck {
		// start health server in the background
		go func() {
			healthRouter := bunrouter.New(
				bunrouter.Use(reqlog.NewMiddleware(
					reqlog.FromEnv("BUNDEBUG"),
				)),
				bunrouter.WithNotFoundHandler(notFoundHandler),
				bunrouter.WithMethodNotAllowedHandler(methodNotAllowedHandler),
			)
			healthRouter.GET(config.HealthCheckPath, func(w http.ResponseWriter, req bunrouter.Request) error { return nil })
			address := fmt.Sprintf(":%d", config.HealthCheckPort)
			logging.Log.WithFields(logrus.Fields{"address": address, "path": config.HealthCheckPath}).Info("starting health server")
			err := http.ListenAndServe(address, healthRouter)
			if err != nil {
				logging.Log.WithError(err).Error("error running health server")
			}
		}()
	}
}

func startServer(config *runCmdConfig) {
	address := fmt.Sprintf(":%d", config.Port)
	router := bunrouter.New(
		bunrouter.Use(reqlog.NewMiddleware(
			reqlog.FromEnv("BUNDEBUG"),
		)),
		bunrouter.WithNotFoundHandler(notFoundHandler),
		bunrouter.WithMethodNotAllowedHandler(methodNotAllowedHandler),
	)

	internal.RegisterRoutes(router)

	logging.Log.WithFields(logrus.Fields{"address": address, "path": "/"}).Info("starting service-go-hello-api server")
	err := http.ListenAndServe(address, router)
	if err != nil {
		logging.Log.WithError(err).Error("error running health server")
	}
}
