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

// type MyCore struct {
// 	zapcore.Core
// }

// func (c *MyCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
// 	// if c.Enabled(entry.Level) {
// 	// 	return checked.AddCore(entry, c)
// 	// }
// 	return checked
// }

// func (c *MyCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
// 	if entry.Level == zapcore.ErrorLevel {
// 		spew.Dump(entry, fields)
// 	}
// 	return c.Core.Write(entry, fields)
// }

func main() {
	logConfig := zap.NewProductionConfig()
	logConfig.OutputPaths = []string{"stdout", "logs/appLog.log"}
	globalLogFields := make(map[string]interface{})
	globalLogFields["service"] = "booksbooksbooks"
	logConfig.InitialFields = globalLogFields

	logger, err := logConfig.Build()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	// fmt.Println("ok")
	// l, err := zap.NewProduction()
	// if err != nil {
	// 	fmt.Println("yerooop")
	// 	panic(err)
	// }
	// logger := zap.New(&MyCore{Core: l.Core()})

	// logger = logger.WithOptions(zap.Hooks(func(log zapcore.Entry) error {
	// 	fmt.Printf("%+v\n", log)
	// 	return nil
	// }))

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
