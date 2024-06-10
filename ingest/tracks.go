package ingest

import (
	"fmt"
	"spotify/api"
	"spotify/models"
	"spotify/utils"

	"github.com/batzz-00/goutils/logger"
)

func (spotify *SpotifyIngest) Tracks() (map[string]api.TopTracksResponse, error) {
	topTrackResp := make(map[string]api.TopTracksResponse)

	for _, period := range spotify.Times {
		logger.Log(fmt.Sprintf("Processing %s_term time range for tracks endpoint", period), logger.Debug)
		tracks, err := spotify.API.TopTracksForUser(period + "_term")
		if err != nil {
			return nil, err
		}

		topTrackResp[period] = tracks
	}
	return topTrackResp, nil
}

func (spotify *SpotifyIngest) PopulateTracks(songs map[string]api.TopTracksResponse, recents api.RecentlyPlayedResponse) ([]models.Song, error) {
	// Songs to attempt to fetch from DB
	songSpotifyIDs := utils.NewStringArgs()
	for _, resp := range songs {
		for _, song := range resp.Items {
			songSpotifyIDs.Add(song.ID)
		}
	}

	for _, song := range recents.Items {
		songSpotifyIDs.Add(song.Track.ID)
	}

	logger.Log(fmt.Sprintf("Querying database for %d songs", len(songSpotifyIDs.Args())), logger.Debug)
	dbSongs, err := spotify.Database.FetchSongsBySpotifyID(songSpotifyIDs.Args())
	if err != nil {
		return nil, err
	}

	// Songs to attempt to fetch from API
	dbSongIds := utils.NewStringArgsFromModel(dbSongs)
	diffedIds := songSpotifyIDs.Diff(dbSongIds)
	// TODO: add unprocessed songs
	songsToFetch := diffedIds.ToString()

	if len(songsToFetch) == 0 {
		logger.Log("Database contains already contains all ingested song references, not querying api", logger.Debug)
		return dbSongs, nil
	}

	apiSongs, err := spotify.API.TracksBySpotifyID(songsToFetch)
	if err != nil {
		return nil, err
	}

	for _, song := range apiSongs {
		newSong := models.NewSong(song.Name, song.ID, song.Album.ID, song.Artists[0].ID, true)
		spotify.OnNewSong(&newSong, true)
		spotify.MetricHandler.AddNewSongIndex(song.ID, song.Name)
		dbSongs = append(dbSongs, newSong)
	}

	return dbSongs, nil
}

func (spotify *SpotifyIngest) AttachTrackUUIDs(songs []models.Song, artists []models.Artist, albums []models.Album) ([]models.Song, error) {
	songValues := []interface{}{}
	for i, song := range songs {
		if !song.NeedsUpdate {
			continue
		}

		for _, artist := range artists {
			if songs[i].ArtistID == artist.SpotifyID {
				songs[i].ArtistID = artist.ID
				break
			}
		}

		if len(songs[i].ArtistID) != 36 {
			logger.Log(fmt.Sprintf("Artist with spotify ID %s was not fetched from spotify", songs[i].ArtistID), logger.Warning)
		}

		for _, album := range albums {
			if songs[i].AlbumID == album.SpotifyID {
				songs[i].AlbumID = album.ID
				break
			}
		}

		songValues = append(songValues, utils.ReflectValues(songs[i])...)
	}

	if len(songValues) == 0 {
		logger.Log("No new song data to ingest", logger.Debug)
		return songs, nil
	}

	songRecords := len(songValues) / len(utils.ReflectColumns(&models.Song{}))
	logger.Log(fmt.Sprintf("Inserting %d new song records", songRecords), logger.Debug)
	err := spotify.Database.Create(&models.Song{}, songValues)
	if err != nil {
		return nil, err
	}

	return songs, nil
}
