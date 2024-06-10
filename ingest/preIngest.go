package ingest

import (
	"database/sql"
	"spotify/api"
	"spotify/models"
	"spotify/utils"

	"github.com/batzz-00/goutils/logger"
)

type PreIngestDatabase interface {
	Create(model models.Model, values []interface{}) error
	FetchUsersBySpotifyIds(names []interface{}) ([]models.User, error)
	FetchArtistsBySpotifyID(spotifyIDs []interface{}) ([]models.Artist, error)
	FetchArtistBySpotifyID(spotifyID string) (models.Artist, error)
}

type PreIngest struct {
	database PreIngestDatabase
	users    []string
}

type IPreIngest interface {
	EnsureBaseDataExists() (string, error)
	GetUserUUID(usernameToReturn string, user api.MeResponse) (string, error)
}

func NewPreIngest(db PreIngestDatabase, users []string) PreIngest {
	return PreIngest{
		users:    users,
		database: db,
	}
}

func (ingest *PreIngest) EnsureBaseDataExists() (string, error) {
	logger.Log("Ensure 'various artists' artist exists", logger.Info)
	dbArtist, err := ingest.database.FetchArtistBySpotifyID(variousArtists)
	if err != nil && err != sql.ErrNoRows {
		return "", nil
	}

	if err == nil {
		logger.Log("Various artists album and artist exists, skipping insert.", logger.Debug)
		return dbArtist.ID, nil
	}

	logger.Log("Inserting 'various artists' artist", logger.Info)
	artist := models.NewArtist("Various artists", variousArtists, false)
	err = ingest.database.Create(&models.Artist{}, utils.ReflectValues(artist))
	if err != nil {
		return "", nil
	}

	return artist.ID, nil
}

func (ingest *PreIngest) GetUserUUID(usernameToReturn string, user api.MeResponse) (string, error) {
	baseUsers := []interface{}{}
	for _, user := range ingest.users {
		baseUsers = append(baseUsers, user)
	}

	users, err := ingest.database.FetchUsersBySpotifyIds(baseUsers)
	if err != nil {
		panic(err)
	}

	userValues := []interface{}{}
Outer:
	for _, username := range baseUsers {
		for _, user := range users {
			if user.SpotifyID == username {
				continue Outer
			}
		}
		newUser := models.NewUser(username.(string), "123", username.(string))
		users = append(users, newUser)
		userValues = append(userValues, utils.ReflectValues(newUser)...)
	}

	userId := ""
	for _, user := range users {
		if user.SpotifyID == usernameToReturn {
			userId = user.ID
		}
	}

	if len(userValues) == 0 {
		logger.Log("Not inserting users, database is already populated.", logger.Debug)
		return userId, nil
	}

	err = ingest.database.Create(&models.User{}, userValues)
	if err != nil {
		return "", err
	}

	return userId, nil
}
