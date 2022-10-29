package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/endpoints"
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
	db.ConnectToRedis()

	logConfig := zap.NewProductionConfig()
	logConfig.OutputPaths = []string{"stdout", "appLog.log"}
	globalLogFields := make(map[string]interface{})
	globalLogFields["service"] = "booksbooksbooks"
	logConfig.InitialFields = globalLogFields

	logger, err := logConfig.Build()
	if err != nil {
		panic(err)
	}

	appConfig := initConfig()
	endpoints.InitConfig(appConfig)
	port := 2945

	router := endpoints.SetupRouter()

	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + fmt.Sprint(port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	logger.Sugar().Infof("Service requests on %d", port)
	log.Fatal(srv.ListenAndServe())
}
