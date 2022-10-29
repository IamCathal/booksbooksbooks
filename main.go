package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/endpoints"
	"github.com/iamcathal/booksbooksbooks/engine"
	"github.com/iamcathal/booksbooksbooks/goodreads"
	"github.com/iamcathal/booksbooksbooks/search"
	"github.com/iamcathal/booksbooksbooks/thebookshop"
	"go.uber.org/zap"
)

var (
	ApplicationStartUpTime time.Time
)

func initConfig() dtos.AppConfig {
	return dtos.AppConfig{
		ApplicationStartUpTime: time.Now(),
	}
}

func main() {
	logConfig := zap.NewProductionConfig()
	logConfig.OutputPaths = []string{"stdout", "appLog.log"}
	globalLogFields := make(map[string]interface{})
	globalLogFields["service"] = "booksbooksbooks"
	logConfig.InitialFields = globalLogFields

	logger, err := logConfig.Build()
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	appConfig := initConfig()
	endpoints.InitConfig(appConfig, logger)
	db.SetLogger(logger)
	engine.SetLogger(logger)
	goodreads.SetLogger(logger)
	thebookshop.SetLogger(logger)
	search.SetLogger(logger)

	port := 2945

	router := endpoints.SetupRouter()
	db.ConnectToRedis()

	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + fmt.Sprint(port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	logger.Sugar().Infof("Service requests on :%d", port)
	log.Fatal(srv.ListenAndServe())
}
