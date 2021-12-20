package main

import (
	"flag"
	"fmt"
	"log"

	"spotify/models"
	"spotify/utils"

	"github.com/batzz-00/goutils/logger"

	"github.com/joho/godotenv"
)

type args struct {
	UserID  string
	Options SpotifyIngestOptions
}

func parseArgs() SpotifyIngestOptions {
	recentListen := flag.Bool("r", false, "Parse and ingest user data regarding a users recently listened tracks")
	topSongs := flag.Bool("t", false, "Parse and ingest user data regarding a users top songs")
	topArtists := flag.Bool("a", false, "Parse and ingest user data regarding a users top artists")
	user := flag.String("u", "", "Username to query the spotify API for, must have relevant refresh_token in env")
	flag.Parse()

	if *user == "" {
		log.Fatalf("UserID must be specified!")
	}

	return SpotifyIngestOptions{
		RecentListen: *recentListen,
		TopSongs:     *topSongs,
		TopArtists:   *topArtists,
		UserID:       *user,
	}
}

func main() {
	args := parseArgs()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	logger.Setup(logger.Debug, nil, logger.NewLoggerOptions("2006-01-02 15:04:05"))

	api := NewSpotifyAPI("https://accounts.spotify.com/", utils.MustGetEnv("secret"), utils.MustGetEnv("clientID"), utils.MustGetEnv(fmt.Sprintf("refresh_%s", args.UserID)), NewAPIOptions(3))

	logger.Log(fmt.Sprintf("Beginning spotify data ingest, user id %s.", args.UserID), logger.Info)

	err = Refresh(api)
	if err != nil {
		logger.Log(err.Error(), logger.Error)
		panic(err)
	}

	database := Database{}
	err = database.Connect()
	if err != nil {
		logger.Log("Failed to connect to database", logger.Error)
		panic(err)
	}

	database.StartTX()

	me, err := api.Me()
	if err != nil {
		logger.Log("Failed to fetch Me endpoint", logger.Error)
		panic(err)
	}

	logger.Log("Handling base user data", logger.Info)
	user, err := HandleBaseUsers(database, args.UserID, me)
	if err != nil {
		logger.Log("Failed when handling base user routine", logger.Error)
		panic(err)
	}
	args.UserID = user.ID

	spotify := newSpotify(&database, api, args)

	err = spotify.Ingest()
	if err != nil {
		database.Rollback()
		panic(err)
	}

	database.Commit()
}

func HandleBaseUsers(db Database, usernameToReturn string, user MeResponse) (models.User, error) {
	baseUsers := []interface{}{"bungusbuster", "anneteresa-gb"}
	users, err := db.FetchUsersByNames(baseUsers)
	if err != nil {
		panic(err)
	}

	userValues := []interface{}{}
	userToReturn := models.User{}
Outer:
	for _, username := range baseUsers {
		for _, user := range users {
			if user.SpotifyID == username {
				continue Outer
			}
		}
		newUser := models.NewUser(username.(string), "123", username.(string))
		users = append(users, newUser)
		userValues = append(userValues, newUser.ToSlice()...)
	}

	for _, user := range users {
		if user.Username == usernameToReturn {
			userToReturn = user
		}
	}

	if len(userValues) == 0 {
		logger.Log("Not inserting users, database is already populated.", logger.Debug)
		return userToReturn, nil
	}

	err = db.CreateUser(userValues)
	if err != nil {
		return models.User{}, err
	}

	return userToReturn, nil
}
