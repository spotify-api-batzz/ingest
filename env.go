package main

import (
	"fmt"
	"log"
	"spotify/api"
	"spotify/database"
	"spotify/metrics"
	"spotify/utils"
	"strings"

	"github.com/joho/godotenv"
)

type SpotifyIngestEnv struct {
	ApiAuth      api.SpotifyAPIAuth
	DbAuth       database.DatabaseAuth
	LogstashAuth metrics.LogstashAuth
	ElasticAuth  metrics.ElasticAuth
	Users        []string
}

func LoadEnv(userID string) SpotifyIngestEnv {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	users := strings.Split(utils.MustGetEnv("users"), ",")

	apiAuth := api.SpotifyAPIAuth{
		Secret:       utils.MustGetEnv("secret"),
		ClientID:     utils.MustGetEnv("clientID"),
		RefreshToken: utils.MustGetEnv(fmt.Sprintf("refresh_%s", userID)),
	}

	logstashAuth := metrics.LogstashAuth{
		Hostname: utils.MustGetEnv("logstash_hostname"),
		Port:     utils.MustGetEnvInt("logstash_port"),
	}

	elasticAuth := metrics.ElasticAuth{
		Username: utils.MustGetEnv("elastic_username"),
		Password: utils.MustGetEnv("elastic_password"),
	}

	dbAuth := database.DatabaseAuth{
		User:     utils.MustGetEnv("DB_USER"),
		IP:       utils.MustGetEnv("DB_IP"),
		Password: utils.MustGetEnv("DB_PASS"),
		Port:     utils.MustGetEnv("DB_PORT"),
		Table:    utils.MustGetEnv("DB_TABLE"),
	}

	return SpotifyIngestEnv{
		ApiAuth:      apiAuth,
		LogstashAuth: logstashAuth,
		ElasticAuth:  elasticAuth,
		DbAuth:       dbAuth,
		Users:        users,
	}
}
