package main

import (
	"fmt"

	"spotify/api"
	"spotify/database"
	"spotify/ingest"
	"spotify/metrics"

	"github.com/batzz-00/goutils/logger"
)

// nice to have:
// retry api requests baked into network client (remove retry from refesh and refreshable interface )
// swap out to a pkg logger
// more integration test cases (some data in db ^ network errors etc ^ calls to event logging)
func main() {
	logger.Setup(logger.Debug, nil, logger.NewLoggerOptions("2006-01-02 15:04:05"))
	args := parseArgs()
	env := LoadEnv(args.UserID)
	args.EnvUsers = env.Users

	database := database.Database{Auth: env.DbAuth}
	err := database.Connect()
	if err != nil {
		logger.Log("Failed to connect to database", logger.Error)
		panic(err)
	}

	database.StartTX()

	ingestContext := ingest.NewIngestContext(args)
	metricHandler, err := metrics.NewMetricHandler(env.LogstashAuth, env.ElasticAuth, ingestContext)
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

	api := api.NewSpotifyAPI("https://accounts.spotify.com/", &metricHandler, env.ApiAuth, api.NewAPIOptions(3))
	logger.Log(fmt.Sprintf("Beginning spotify data ingest, user id %s.", args.UserID), logger.Info)

	err = Refresh(&api)
	if err != nil {
		logger.Log(err.Error(), logger.Error)
		metricHandler.AddNewFailure("REFRESH_TOKEN", err)
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
