package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"spotify/logger"
	"spotify/models"
	"spotify/utils"

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

	api := NewAPI("https://accounts.spotify.com/", utils.MustGetEnv("secret"), utils.MustGetEnv("clientID"), utils.MustGetEnv(fmt.Sprintf("refresh_%s", os.Args[1])))

	err = api.Refresh()
	if err != nil {
		logger.Log("Failed to refresh API token!", logger.Error)
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

	user, err := HandleBaseUsers(database, os.Args[1], me)
	if err != nil {
		logger.Log("Failed when handling base user routine", logger.Error)
		panic(err)
	}

	spotify := newSpotify(&database, &api, user.ID)

	err = spotify.DataInserts()
	if err != nil {
		panic(err)
	}
}

func HandleBaseUsers(db Database, usernameToReturn string, user MeResponse) (models.User, error) {
	logger.Log("Handling base user data", logger.Notice)
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
		logger.Log("syke bitch the database already contains all necessary users", logger.Notice)
		return userToReturn, nil
	}

	err = db.CreateUser(userValues)
	if err != nil {
		return models.User{}, err
	}

	return userToReturn, nil
}

func BasicAuth(clientID string, clientSecret string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret))))
}
