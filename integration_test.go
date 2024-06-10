package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"spotify/api"
	"spotify/database"
	"spotify/ingest"
	"spotify/metrics"
	"spotify/models"
	"spotify/utils"
	"testing"

	differ "github.com/andreyvit/diff"

	"github.com/batzz-00/goutils/logger"
	"github.com/go-test/deep"
)

// RUN WITH BUILD TAGs
// go test . -tags=test
func TestIntegration(t *testing.T) {
	args := ingest.SpotifyIngestOptions{
		RecentListen:       true,
		TopSongs:           false,
		TopArtists:         false,
		UserID:             "123",
		VariousArtistsUUID: "123",
		EnvUsers:           []string{"123"},
	}

	logger.Setup(logger.Debug, nil, logger.NewLoggerOptions("2006-01-02 15:04:05"))
	db := database.NewMockDatabase()
	api := api.NewMockSpotifyApi("recent-listens")
	logger.Log(fmt.Sprintf("Beginning spotify data ingest, user id %s.", args.UserID), logger.Info)

	err := Refresh(&api)
	if err != nil {
		logger.Log(err.Error(), logger.Error)
		t.Fatal(err)
	}

	metricHander := metrics.NewMockMetricHandler()

	preingest := ingest.NewPreIngest(&db, []string{"123"})
	spotify := ingest.BootstrapSpotifyingest(&db, &api, args, &preingest, &metricHander)
	err = spotify.Ingest()
	if err != nil {
		t.Fatal(err)
	}

	modelSlice := []models.Model{&models.Song{}, &models.Artist{}, &models.RecentListen{}, &models.Thumbnail{}, &models.User{}, &models.Album{}}
	expectedInserts := loadExpectedInserts("recent-listens", modelSlice)

	dbBytes, _ := json.Marshal(db.SavedValues)
	expectedBytes, _ := json.Marshal(expectedInserts)

	dbString := string(dbBytes)
	expectedString := string(expectedBytes)

	if diff := deep.Equal(dbString, expectedString); diff != nil {
		// TODO: sane diff printing pls
		t.Errorf(differ.LineDiff(dbString, expectedString))
	}

}

func loadExpectedInserts(test string, models []models.Model) map[string][]interface{} {
	loader := utils.LoadJSON("expected", test)
	expected := make(map[string][]interface{})

	for _, model := range models {
		structName := reflect.TypeOf(model).Elem().Name()
		bytes := loader(fmt.Sprintf("%s-insert", structName))

		var values []interface{}
		json.Unmarshal(bytes, &values)

		expected[structName] = values
	}

	return expected
}
