package ingest

import (
	"fmt"
	"spotify/models"
	"spotify/types"
	"spotify/utils"

	"github.com/batzz-00/goutils/logger"
)

func (spotify *SpotifyIngest) PopulateArtists(songs map[string]types.TopTracksResponse, artists map[string]types.TopArtistsResponse, recents types.RecentlyPlayedResponse) ([]models.Artist, error) {
	// Songs to attempt to fetch from DB

	artistSpotifyIDs := utils.NewStringArgs()
	for _, resp := range songs {
		for _, song := range resp.Items {
			for _, artist := range song.Artists {
				artistSpotifyIDs.Add(artist.ID)
			}
			for _, artist := range song.Album.Artists {
				artistSpotifyIDs.Add(artist.ID)
			}
		}
	}

	for _, resp := range artists {
		for _, artist := range resp.Items {
			artistSpotifyIDs.Add(artist.ID)
		}
	}

	for _, song := range recents.Items {
		for _, artist := range song.Track.Album.Artists {
			artistSpotifyIDs.Add(artist.ID)
		}
		for _, artist := range song.Track.Artists {
			artistSpotifyIDs.Add(artist.ID)
		}
	}

	logger.Log(fmt.Sprintf("Querying database for %d artists", len(artistSpotifyIDs.Args())), logger.Debug)
	dbArtists, err := spotify.Database.FetchArtistsBySpotifyID(artistSpotifyIDs.Args())
	if err != nil {
		return nil, err
	}

	// Songs to attempt to fetch from API
	artistsToFetch := []string{}
Outer:
	for id := range artistSpotifyIDs.UniqueMap {
		for _, dbArtist := range dbArtists {
			if dbArtist.SpotifyID == id {
				logger.Log(fmt.Sprintf("Database already contains artist %s (%s)", dbArtist.Name, dbArtist.SpotifyID), logger.Trace)
				continue Outer
			}
		}
		artistsToFetch = append(artistsToFetch, id)
	}

	if len(artistsToFetch) == 0 {
		logger.Log("Database contains already contains all ingested artist references, not querying api", logger.Debug)
		return dbArtists, nil
	}

	apiArtists, err := spotify.API.ArtistsBySpotifyID(artistsToFetch)
	if err != nil {
		return nil, err
	}

	for _, artist := range apiArtists {
		dbArtists = append(dbArtists, models.NewArtist(artist.Name, artist.ID, true))
	}

	return dbArtists, nil
}

func (spotify *SpotifyIngest) Artists() (map[string]types.TopArtistsResponse, error) {
	artistsResp := make(map[string]types.TopArtistsResponse)

	for _, period := range spotify.Times {
		logger.Log(fmt.Sprintf("Processing %s_term time range for artists endpoint", period), logger.Debug)
		artists, err := spotify.API.TopArtistsForUser(period + "_term")
		if err != nil {
			return nil, err
		}
		artistsResp[period] = artists
	}
	return artistsResp, nil
}

func (spotify *SpotifyIngest) InsertArtists(artists []models.Artist) error {
	artistValues := []interface{}{}
	for _, artist := range artists {
		if !artist.NeedsUpdate {
			continue
		}
		artistValues = append(artistValues, utils.ReflectValues(artist)...)
	}

	if len(artistValues) == 0 {
		logger.Log("No new artist data to ingest", logger.Debug)
		return nil
	}

	logger.Log("Inserting new artists", logger.Debug)
	err := spotify.Database.Create(&models.Artist{}, artistValues)
	if err != nil {
		return err
	}

	return nil
}
