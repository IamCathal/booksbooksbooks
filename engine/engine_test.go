package engine

import (
	"log"
	"os"
	"testing"

	"go.uber.org/zap"
)

var (
	validShelfURL = "https://www.goodreads.com/review/list/26367680?shelf=read"
)

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	SetLogger(logger)

	code := m.Run()

	os.Exit(code)
}

// func TestAutomatedCheckEngineCorrectlyChecksCurrentTime(t *testing.T) {
// 	mockController := &controller.MockCntrInterface{}

// 	currTime := "18:30"
// 	mockController.On("GetFormattedTime").Return(currTime)

// 	db.SetAutomatedBookShelfCrawlTime("19:30")

// 	mockController.Ass
// }

func TestWorker(t *testing.T) {

}
