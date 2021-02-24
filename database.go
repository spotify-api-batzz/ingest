package main

import (
	"fmt"
	"os"
	"spotify/logger"
	"spotify/models"
	"strings"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

type Database struct {
	DB *sqlx.DB
}

// Connect opens up a conection to the database
func (database *Database) Connect() error {
	dbUser := os.Getenv("DB_USER")
	dbIP := os.Getenv("DB_IP")
	dbPass := os.Getenv("DB_PASS")
	dbPort := os.Getenv("DB_PORT")
	dbTable := os.Getenv("DB_TABLE")

	url := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s", dbIP, dbPort, dbTable, dbUser, dbPass)

	db, err := sqlx.Connect("pgx", url)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		logger.Log(err, logger.Error)
		// do something here
		return err
	}

	database.DB = db
	return nil
}

func (d *Database) FetchArtistBySpotifyID(spotifyID string) (models.Artist, error) {
	artist := models.Artist{}
	err := d.DB.Get(&artist, "SELECT * FROM artists WHERE spotify_id = $1", spotifyID)
	if err != nil {
		return models.Artist{}, err
	}
	return artist, nil
}

func (d *Database) FetchSongsBySpotifyID(spotifyIDs []interface{}) ([]models.Song, error) {
	songs := []models.Song{}
	sql := fmt.Sprintf("SELECT * FROM songs WHERE spotify_id IN (%s)", PrepareBatchValuesPG(1, len(spotifyIDs)))
	err := d.DB.Select(&songs, sql, spotifyIDs...)
	if err != nil {
		return nil, err
	}
	return songs, nil
}

func (d *Database) FetchAlbumsBySpotifyID(spotifyIDs []interface{}) ([]models.Album, error) {
	albums := []models.Album{}
	sql := fmt.Sprintf("SELECT * FROM albums WHERE spotify_id IN (%s)", PrepareBatchValuesPG(1, len(spotifyIDs)))
	err := d.DB.Select(&albums, sql, spotifyIDs...)
	if err != nil {
		return nil, err
	}
	return albums, nil
}

func (d *Database) FetchArtistsBySpotifyID(spotifyIDs []interface{}) ([]models.Artist, error) {
	artists := []models.Artist{}
	sql := fmt.Sprintf("SELECT * FROM artists WHERE spotify_id IN (%s)", PrepareBatchValuesPG(1, len(spotifyIDs)))
	err := d.DB.Select(&artists, sql, spotifyIDs...)
	if err != nil {
		return nil, err
	}
	return artists, nil
}

func (d *Database) FetchArtistByID(id string) (models.Artist, error) {
	artist := models.Artist{}
	err := d.DB.Get(&artist, "SELECT * FROM artists WHERE id = $1", id)
	if err != nil {
		return models.Artist{}, err
	}
	return artist, nil
}

func (d *Database) CreateArtist(artistValues []interface{}) error {
	sql := fmt.Sprintf("INSERT INTO artists (id, name, spotify_id, created_at, updated_at) VALUES %s ", PrepareBatchValuesPG(5, len(artistValues)/5))
	_, err := d.DB.Exec(sql, artistValues...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateTopArtist(topArtistValues []interface{}) error {
	sql := fmt.Sprintf(`INSERT INTO top_artists (id, artist_id, "order", time_period, created_at, updated_at) VALUES %s `, PrepareBatchValuesPG(6, len(topArtistValues)/6))
	_, err := d.DB.Exec(sql, topArtistValues...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateSong(songValues []interface{}) error {
	sql := fmt.Sprintf("INSERT INTO songs (id, spotify_id, album_id, artist_id, name, created_at, updated_at) VALUES %s ", PrepareBatchValuesPG(7, len(songValues)/7))
	_, err := d.DB.Exec(sql, songValues...)
	if err != nil {
		return err
	}
	return nil
}

// slice := make([]interface{}, 6)
// slice[0] = t.ID
// slice[1] = t.Entity
// slice[2] = t.EntityID
// slice[3] = t.URL
// slice[4] = time.Now().UTC().Format(time.RFC3339)
// slice[5] = time.Now().UTC().Format(time.RFC3339)

// return slice

func (d *Database) CreateThumbnail(thumbnailValues []interface{}) error {
	sql := fmt.Sprintf("INSERT INTO thumbnails (id, entity, entity_id, url, created_at, updated_at) VALUES %s ", PrepareBatchValuesPG(6, len(thumbnailValues)/6))
	_, err := d.DB.Exec(sql, thumbnailValues...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateAlbum(albumValues []interface{}) error {
	sql := fmt.Sprintf("INSERT INTO albums (id, name, artist_id, spotify_id, created_at, updated_at) VALUES %s ", PrepareBatchValuesPG(6, len(albumValues)/6))
	_, err := d.DB.Exec(sql, albumValues...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateTopSong(topArtistValues []interface{}) error {
	sql := fmt.Sprintf(`INSERT INTO top_songs (id, user_id, song_id, "order", time_period, created_at, updated_at) VALUES %s `, PrepareBatchValuesPG(7, len(topArtistValues)/7))
	_, err := d.DB.Exec(sql, topArtistValues...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateRecentlyListened(recentlyListenedValues []interface{}) error {
	sql := fmt.Sprintf(`INSERT INTO recent_listens (id, song_id, user_id, played_at, created_at, updated_at) VALUES %s `, PrepareBatchValuesPG(6, len(recentlyListenedValues)/6))
	_, err := d.DB.Exec(sql, recentlyListenedValues...)
	if err != nil {
		return err
	}
	return nil
}

func PrepareBatchValuesPG(paramLength int, valueLength int) string {
	counter := 1
	var values string
	for i := 0; i < valueLength; i++ {
		values = fmt.Sprintf("%s, %s", values, genValString(paramLength, &counter))
	}
	return strings.TrimPrefix(values, ", ")
}

func genValString(paramLength int, counter *int) string {
	var valString string
	for i := 0; i < paramLength; i++ {
		valString = valString + fmt.Sprintf("$%d,", *counter)
		*counter++
	}
	valString = fmt.Sprintf("(%s)", strings.TrimSuffix(valString, ","))
	return valString
}
