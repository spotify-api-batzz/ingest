package ingest

import (
	"fmt"
	"spotify/api"
	"spotify/models"
	"spotify/utils"
	"time"

	"github.com/batzz-00/goutils/logger"
)

func (spotify *SpotifyIngest) InsertRecentListens(recents api.RecentlyPlayedResponse, songs []models.Song, existingRecentListens []models.RecentListen) error {
	recentListenValues := []interface{}{}
Outer:
	for _, recentListen := range recents.Items {
		for _, existingRecentListen := range existingRecentListens {
			if recentListen.PlayedAt.Format(time.RFC3339) == existingRecentListen.PlayedAt.Format(time.RFC3339) {
				continue Outer
			}
		}

		newRecentListenData := models.NewRecentListen("", spotify.Options.UserID, recentListen.PlayedAt)
		// TODO: move to an attach song uuid list
		spotify.OnNewEntityEvent(&newRecentListenData)
		song, exists := getSongBySpotifyID(songs, recentListen.Track.ID)
		if exists {
			newRecentListenData.SongID = song.ID
		} else {
			fmt.Printf("Failed to find %s\n", recentListen.Track.ID)
		}
		recentListenValues = append(recentListenValues, utils.ReflectValues(newRecentListenData)...)
	}

	if len(recentListenValues) == 0 {
		logger.Log("No recent listen data to ingest", logger.Info)
		return nil
	}

	logger.Log("Inserting new recently listened to songs", logger.Info)
	recentListenRecords := len(recentListenValues) / len(utils.ReflectColumns(&models.RecentListen{}))
	logger.Log(fmt.Sprintf("Inserting %d new recent_listen records", recentListenRecords), logger.Debug)
	err := spotify.Database.Create(&models.RecentListen{}, recentListenValues)
	if err != nil {
		return err
	}

	return nil
}

func (spotify *SpotifyIngest) FetchExistingRecentListens(recents api.RecentlyPlayedResponse) ([]models.RecentListen, error) {
	recentPlayedAtList := []interface{}{}
	var earliestRecentlyPlayedAt time.Time
	for _, recent := range recents.Items {
		recentPlayedAtList = append(recentPlayedAtList, recent.PlayedAt.Format(time.RFC3339))
		if earliestRecentlyPlayedAt.IsZero() {
			earliestRecentlyPlayedAt = recent.PlayedAt
		} else if earliestRecentlyPlayedAt.After(recent.PlayedAt) {
			earliestRecentlyPlayedAt = recent.PlayedAt
		}
	}

	if len(recentPlayedAtList) == 0 {
		logger.Log("User has no recently played songs", logger.Debug)
		return []models.RecentListen{}, nil
	}

	recentListens, err := spotify.Database.FetchRecentListensByUserIDAndTime(spotify.Options.UserID, recentPlayedAtList, earliestRecentlyPlayedAt.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}

	return recentListens, nil
}
