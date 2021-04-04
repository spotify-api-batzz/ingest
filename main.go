package main

import (
	"fmt"
	"log"
	"os"

	"spotify/models"
	"spotify/utils"

	"github.com/batzz-00/goutils/logger"

	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) == 1 {
		panic("You must specify a userID for your first command line arg!")
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	logger.Setup(logger.Info, nil, logger.NewLoggerOptions("2006-01-02 15:04:05"))

	api := NewSpotifyAPI("https://accounts.spotify.com/", utils.MustGetEnv("secret"), utils.MustGetEnv("clientID"), utils.MustGetEnv(fmt.Sprintf("refresh_%s", os.Args[1])), NewAPIOptions(3))

	logger.Log(fmt.Sprintf("Beggining spotify data ingest, user id %s.", os.Args[1]), logger.Info)

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

	me, err := api.Me()
	if err != nil {
		logger.Log("Failed to fetch Me endpoint", logger.Error)
		panic(err)
	}

	logger.Log("Handling base user data", logger.Info)
	user, err := HandleBaseUsers(database, os.Args[1], me)
	if err != nil {
		logger.Log("Failed when handling base user routine", logger.Error)
		panic(err)
	}

	spotify := newSpotify(&database, api, user.ID)

	err = spotify.DataInserts()
	if err != nil {
		panic(err)
	}
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
			if user.Username == usernameToReturn {
				userToReturn = user
			}
			if user.SpotifyID == username {
				continue Outer
			}
		}
		newUser := models.NewUser(username.(string), "123", username.(string))
		userValues = append(userValues, newUser.ToSlice()...)
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
