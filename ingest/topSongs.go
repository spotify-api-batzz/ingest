package ingest

import (
	"fmt"
	"spotify/api"
	"spotify/models"
	"spotify/utils"

	"github.com/batzz-00/goutils/logger"
)

func (spotify *SpotifyIngest) InsertTopSongs(songs map[string]api.TopTracksResponse, dbSongs []models.Song) error {
	topSongDataValues := []interface{}{}
	topSongValues := []interface{}{}

	topSong := models.NewTopSong(spotify.Options.UserID)
	spotify.OnNewEntityEvent(&topSong)
	for term, resp := range songs {
		for i, song := range resp.Items {
			dbSong, exists := getSongBySpotifyID(dbSongs, song.ID)
			newTopSongData := models.NewTopSongData(topSong.ID, "", i+1, term)
			spotify.OnNewEntityEvent(&newTopSongData)
			if exists {
				newTopSongData.SongID = dbSong.ID
			} else {
				logger.Log(fmt.Sprintf("Failed to attach song ID for song %s", song.Name), logger.Warning)
			}

			topSongDataValues = append(topSongDataValues, utils.ReflectValues(newTopSongData)...)
		}
	}

	topSongValues = append(topSongValues, utils.ReflectValues(topSong)...)

	if len(topSongDataValues) == 0 {
		logger.Log("No top song data to ingest.", logger.Info)
		return nil
	}

	topSongRecordsToInsert := len(topSongValues) / len(utils.ReflectColumns(topSong))
	logger.Log("Inserting new top song record", logger.Info)
	logger.Log(fmt.Sprintf("Inserting %d top_song records", topSongRecordsToInsert), logger.Debug)
	err := spotify.Database.Create(&models.TopSong{}, topSongValues)
	if err != nil {
		return err
	}

	topSongDataRecordsToInsert := len(topSongDataValues) / len(utils.ReflectColumns(models.TopSongData{}))
	logger.Log("Inserting new top song data", logger.Info)
	logger.Log(fmt.Sprintf("Inserting %d top_song_data records", topSongDataRecordsToInsert), logger.Debug)
	err = spotify.Database.Create(&models.TopSongData{}, topSongDataValues)
	if err != nil {
		return err
	}

	return nil
}
