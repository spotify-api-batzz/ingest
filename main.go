package main

import (
	"fmt"
	"log"
	"strings"

	"spotify/api"
	"spotify/database"
	"spotify/ingest"
	"spotify/metrics"
	"spotify/utils"

	"github.com/batzz-00/goutils/logger"

	"github.com/joho/godotenv"
)

// func()

// nice to have:
// retry api requests baked into network client (remove retry from refesh and refreshable interface )
// swap out to a pkg logger
// more integration test cases (some data in db ^ network errors etc ^ calls to event logging)
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	args := parseArgs()
	args.EnvUsers = strings.Split(utils.MustGetEnv("users"), ",")

	logger.Setup(logger.Debug, nil, logger.NewLoggerOptions("2006-01-02 15:04:05"))

	database := database.Database{}
	err = database.Connect()
	if err != nil {
		logger.Log("Failed to connect to database", logger.Error)
		panic(err)
	}

	database.StartTX()

	spotifyAPIAuth := api.SpotifyAPIAuth{
		Secret:       utils.MustGetEnv("secret"),
		ClientID:     utils.MustGetEnv("clientID"),
		RefreshToken: utils.MustGetEnv(fmt.Sprintf("refresh_%s", args.UserID)),
	}

	ingestContext := ingest.NewIngestContext(args)
	logStashHostname := utils.MustGetEnv("logstash_hostname")
	logStashPort := utils.MustGetEnvInt("logstash_port")
	metricHandler, err := metrics.NewMetricHandler(logStashHostname, logStashPort, ingestContext)
	if err != nil {
		logger.Log("Failed to make metrics handler", logger.Error)
		panic(err)
	}

	var addIngestFinishedIndex = metricHandler.AddIngestFinishedIndex
	var addOnNewEntityIndex = metricHandler.AddNewModel

	args.Events = ingest.SpotifyIngestEvents{
		OnNewEntity: &addOnNewEntityIndex,
		OnFinish:    &addIngestFinishedIndex,
	}

	api := api.NewSpotifyAPI("https://accounts.spotify.com/", &metricHandler, spotifyAPIAuth, api.NewAPIOptions(3))
	logger.Log(fmt.Sprintf("Beginning spotify data ingest, user id %s.", args.UserID), logger.Info)

	err = Refresh(&api)
	if err != nil {
		logger.Log(err.Error(), logger.Error)
		metricHandler.AddNewFailure("REFRESH_API", err)
		panic(err)
	}

	preingest := ingest.NewPreIngest(&database, args.EnvUsers)
	spotify := ingest.BootstrapSpotifyingest(&database, &api, &preingest, args)
	err = spotify.Ingest()
	if err != nil {
		database.Rollback()
		metricHandler.AddNewFailure("INGEST", err)
		panic(err)
	}

	err = metricHandler.Close()
	if err != nil {
		database.Rollback()
		panic(err)
	}

	database.Commit()
}
