package util

import (
	"log"
	"os"
	"testing"

	"github.com/iamcathal/booksbooksbooks/db"
	"go.uber.org/zap"
	"gotest.tools/assert"
)

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	db.SetLogger(logger)

	code := m.Run()

	os.Exit(code)
}

func TestIsBookEnglishDetectFrenchDune(t *testing.T) {
	assert.Equal(t, false, IsEnglishText("Herbert, Frank - Le Messie de Dune ( FRENCH LANGUAGE PB ED) - En Francais"))
}

func TestIsBookEnglishDetectsFrenchBook(t *testing.T) {
	assert.Equal(t, false, IsEnglishText("Purdy, James - Les Enfants , C'est tout - PB Gallimard - 1968 "))
}

func TestIsBookEnglishDetectsSpanishBook(t *testing.T) {
	assert.Equal(t, false, IsEnglishText("Auel, Jean M. - El Clan Del Oso Cavernario ( Spanish Edition)"))
}

func TestIsBookEnglishDetectsBookWithFrenchFada(t *testing.T) {
	assert.Equal(t, false, IsEnglishText("Parrot, André - Sumer - FRENCH LANGUAGE Edition"))
}

func TestIsBookEnglishDetectsBookWithAUmlaut(t *testing.T) {
	assert.Equal(t, false, IsEnglishText("Doerr, Anthony -Kaikki se valo jota emme näe - HB - Finnish"))
}
func TestIsBookEnglishDetectsBookWithPolishFancyZ(t *testing.T) {
	assert.Equal(t, false, IsEnglishText("McCaffrey, Anne -Historia Nerilki ( Jeźdźcy smoków z"))
}
func TestIsBookEnglishDetectsBookWithCyrillicLetters(t *testing.T) {
	assert.Equal(t, false, IsEnglishText("имя ветра"))
}
func TestIsBookEnglishDetectsBookWithFSharphesS(t *testing.T) {
	assert.Equal(t, false, IsEnglishText("Schiller, Friedrich - Geschichte des dreißigjährigen Kriegs"))
}
func TestIsBookEnglishDetectsEnglishTitleBook(t *testing.T) {
	assert.Equal(t, true, IsEnglishText("Collins, Suzanne / The Hunger Games ( Hunger Games Trilogy "))
}
