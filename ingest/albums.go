package ingest

import (
	"fmt"
	"spotify/api"
	"spotify/models"
	"spotify/utils"

	"github.com/batzz-00/goutils/logger"
)

func (spotify *SpotifyIngest) PopulateAlbums(songs map[string]api.TopTracksResponse, recents api.RecentlyPlayedResponse) ([]models.Album, error) {
	albumSpotifyIDs := utils.NewStringArgs()
	for _, resp := range songs {
		for _, song := range resp.Items {
			albumSpotifyIDs.Add(song.Album.ID)
		}
	}

	for _, song := range recents.Items {
		albumSpotifyIDs.Add(song.Track.Album.ID)
	}

	logger.Log(fmt.Sprintf("Querying database for %d albums", len(albumSpotifyIDs.Args())), logger.Debug)
	dbAlbums, err := spotify.Database.FetchAlbumsBySpotifyID(albumSpotifyIDs.Args())
	if err != nil {
		return nil, err
	}

	// Songs to attempt to fetch from API
	albumsToFetch := []string{}
Outer:
	for id := range albumSpotifyIDs.UniqueMap {
		for _, dbAlbum := range dbAlbums {
			if dbAlbum.SpotifyID == id {
				continue Outer
			}
		}
		albumsToFetch = append(albumsToFetch, id)
	}

	if len(albumsToFetch) == 0 {
		logger.Log("Database contains already contains all ingested album references, not querying api", logger.Debug)
		return dbAlbums, nil
	}

	apiAlbums, err := spotify.API.AlbumsBySpotifyID(albumsToFetch)
	if err != nil {
		return nil, err
	}

	for _, album := range apiAlbums {
		album := models.NewAlbum(album.Name, album.ID, album.Artists[0].ID, true)
		dbAlbums = append(dbAlbums, album)
		spotify.OnNewAlbum(&album, true)
		spotify.MetricHandler.AddNewAlbumIndex(album.ID, album.Name)
	}

	return dbAlbums, nil
}

func (spotify *SpotifyIngest) AttachAlbumUUIDs(albums []models.Album, artists []models.Artist) error {
	albumValues := []interface{}{}

	for i, album := range albums {
		if !album.NeedsUpdate {
			continue
		}
		for _, artist := range artists {
			// hardcoded 'various artists'
			if albums[i].ArtistID == "0LyfQWJT6nXafLPZqxe9Of" {
				albums[i].ArtistID = spotify.Options.VariousArtistsUUID
			}
			if albums[i].ArtistID == artist.SpotifyID {
				albums[i].ArtistID = artist.ID
				break
			}
		}
		albumValues = append(albumValues, utils.ReflectValues(albums[i])...)
	}

	if len(albumValues) == 0 {
		logger.Log("No new album data to ingest", logger.Debug)
		return nil
	}

	logger.Log("Inserting new albums", logger.Debug)
	err := spotify.Database.Create(&models.Album{}, albumValues)
	if err != nil {
		return err
	}

	return nil
}
