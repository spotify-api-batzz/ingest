package ingest

import (
	"fmt"
	"spotify/api"
	"spotify/models"
	"spotify/utils"

	"github.com/batzz-00/goutils/logger"
)

func (spotify *SpotifyIngest) InsertTopArtists(topArtists map[string]api.TopArtistsResponse, dbArtists []models.Artist) error {
	topArtistDataValues := []interface{}{}
	topArtistValues := []interface{}{}
	newTopArtist := models.NewTopArtist(spotify.Options.UserID)
	spotify.OnNewEntityEvent(&newTopArtist)

	for term, resp := range topArtists {
		for i, artist := range resp.Items {
			dbArtist, exists := getArtistBySpotifyID(dbArtists, artist.ID)
			newTopArtistData := models.NewTopArtistData(artist.Name, "", i+1, term, newTopArtist.ID)
			spotify.OnNewEntityEvent(&newTopArtistData)
			if exists {
				newTopArtistData.ArtistID = dbArtist.ID
			} else {
				logger.Log(fmt.Sprintf("Failed to attach artist ID for artist %s", artist.Name), logger.Warning)
			}
			topArtistDataValues = append(topArtistDataValues, utils.ReflectValues(newTopArtistData)...)
		}
	}

	topArtistValues = append(topArtistValues, utils.ReflectValues(newTopArtist)...)
	if len(topArtistDataValues) == 0 {
		logger.Log("No top artist data to ingest.", logger.Info)
		return nil
	}

	logger.Log("Inserting new top artist record", logger.Info)
	topArtistRecords := len(topArtistValues) / len(utils.ReflectColumns(&models.TopArtist{}))
	logger.Log(fmt.Sprintf("Inserting %d top_artist records", topArtistRecords), logger.Debug)
	err := spotify.Database.Create(&models.TopArtist{}, topArtistValues)
	if err != nil {
		return err
	}

	logger.Log("Inserting new top artist data", logger.Info)
	topArtistDataRecords := len(topArtistDataValues) / len(utils.ReflectColumns(&models.TopArtistData{}))
	logger.Log(fmt.Sprintf("Inserting %d top_artist_data records", topArtistDataRecords), logger.Debug)
	err = spotify.Database.Create(&models.TopArtistData{}, topArtistDataValues)
	if err != nil {
		return err
	}

	return nil
}
