package main

import (
	"context"
	"fmt"
	"spotify/models"
	"spotify/utils"
	"strings"

	"github.com/batzz-00/goutils/logger"

	"github.com/batzz-00/goutils"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

type Database struct {
	DB *sqlx.DB
	Tx *sqlx.Tx
}

func (database *Database) StartTX() {
	if database.Tx == nil {
		database.Tx = database.DB.MustBeginTx(context.Background(), nil)
	}
}

func (database *Database) Rollback() {
	if database.Tx != nil {
		logger.Log("Rolling back changes", logger.Info)
		database.Tx.Rollback()
		return
	}
	logger.Log("No transaction instance to rollback!", logger.Warning)
}

func (database *Database) Commit() {
	if database.Tx != nil {
		logger.Log("Committing changes", logger.Info)
		database.Tx.Commit()
		return
	}
	logger.Log("No transaction instance to commit!", logger.Warning)
}

// Connect opens up a conection to the database
func (database *Database) Connect() error {
	dbUser := utils.MustGetEnv("DB_USER")
	dbIP := utils.MustGetEnv("DB_IP")
	dbPass := utils.MustGetEnv("DB_PASS")
	dbPort := utils.MustGetEnv("DB_PORT")
	dbTable := utils.MustGetEnv("DB_TABLE")

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

func (d *Database) MustGetTx() *sqlx.Tx {
	if d.Tx == nil {
		d.StartTX()
	}

	return d.Tx
}

func (d *Database) FetchUsersBySpotifyIds(names []interface{}) ([]models.User, error) {
	user := []models.User{}
	sql := fmt.Sprintf("SELECT * FROM users WHERE spotify_id IN (%s)", PrepareInStringPG(1, len(names), 1))
	err := d.MustGetTx().Select(&user, sql, names...)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *Database) FetchUserByName(name string) (models.User, error) {
	user := models.User{}
	err := d.MustGetTx().Get(&user, "SELECT * FROM users WHERE username = $1", name)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (d *Database) FetchArtistBySpotifyID(spotifyID string) (models.Artist, error) {
	artist := models.Artist{}
	err := d.MustGetTx().Get(&artist, "SELECT * FROM artists WHERE spotify_id = $1", spotifyID)
	if err != nil {
		return models.Artist{}, err
	}
	return artist, nil
}

func (d *Database) FetchSongsBySpotifyID(spotifyIDs []interface{}) ([]models.Song, error) {
	songs := []models.Song{}
	sql := fmt.Sprintf("SELECT * FROM songs WHERE spotify_id IN (%s)", PrepareBatchValuesPG(1, len(spotifyIDs)))
	err := d.MustGetTx().Select(&songs, sql, spotifyIDs...)
	if err != nil {
		return nil, err
	}
	return songs, nil
}

func (d *Database) FetchAlbumsBySpotifyID(spotifyIDs []interface{}) ([]models.Album, error) {
	albums := []models.Album{}
	sql := fmt.Sprintf("SELECT * FROM albums WHERE spotify_id IN (%s)", PrepareBatchValuesPG(1, len(spotifyIDs)))
	err := d.MustGetTx().Select(&albums, sql, spotifyIDs...)
	if err != nil {
		return nil, err
	}
	return albums, nil
}

func (d *Database) FetchArtistsBySpotifyID(spotifyIDs []interface{}) ([]models.Artist, error) {
	artists := []models.Artist{}
	sql := fmt.Sprintf("SELECT * FROM artists WHERE spotify_id IN (%s)", PrepareBatchValuesPG(1, len(spotifyIDs)))
	err := d.MustGetTx().Select(&artists, sql, spotifyIDs...)
	if err != nil {
		return nil, err
	}
	return artists, nil
}

func (d *Database) FetchArtistByID(id string) (models.Artist, error) {
	artist := models.Artist{}
	err := d.MustGetTx().Get(&artist, "SELECT * FROM artists WHERE id = $1", id)
	if err != nil {
		return models.Artist{}, err
	}
	return artist, nil
}

// earliest time optimization
func (d *Database) FetchRecentListensByUserIDAndTime(userID string, recentListenedToIDs []interface{}, earliestTime interface{}) ([]models.RecentListen, error) {
	recentListens := []models.RecentListen{}
	columnNames := goutils.ColumnNamesExclusive(&models.RecentListen{})
	tableName := (&models.RecentListen{}).TableName()
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 AND played_at >= $2 AND played_at IN (%s)", columnNames, tableName, PrepareInStringPG(1, len(recentListenedToIDs), 3))
	vars := []interface{}{userID, earliestTime}
	vars = append(vars, recentListenedToIDs...)
	err := d.MustGetTx().Select(&recentListens, sql, vars...)
	if err != nil {
		return nil, err
	}
	return recentListens, nil
}

func (d *Database) FetchThumbnailsByEntityID(entityIDs []interface{}) ([]models.Thumbnail, error) {
	thumbnails := []models.Thumbnail{}
	sql := fmt.Sprintf("SELECT * FROM thumbnails WHERE entity_id IN (%s)", PrepareInStringPG(1, len(entityIDs), 1))
	err := d.MustGetTx().Select(&thumbnails, sql, entityIDs...)
	if err != nil {
		return nil, err
	}
	return thumbnails, nil
}

func (d *Database) CreateArtist(artistValues []interface{}) error {
	sql := fmt.Sprintf("INSERT INTO artists (id, name, spotify_id, created_at, updated_at) VALUES %s ", PrepareBatchValuesPG(5, len(artistValues)/5))
	_, err := d.MustGetTx().Exec(sql, artistValues...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateUser(userValues []interface{}) error {
	sql := fmt.Sprintf("INSERT INTO users (id, username, password, spotify_id, created_at, updated_at) VALUES %s ", PrepareBatchValuesPG(6, len(userValues)/6))
	_, err := d.MustGetTx().Exec(sql, userValues...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateTopArtist(topArtistValues []interface{}) error {
	sql := fmt.Sprintf(`INSERT INTO top_artists (id, artist_id, "order", user_id, time_period, created_at, updated_at) VALUES %s `, PrepareBatchValuesPG(7, len(topArtistValues)/7))
	_, err := d.MustGetTx().Exec(sql, topArtistValues...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateSong(songValues []interface{}) error {
	sql := fmt.Sprintf("INSERT INTO songs (id, spotify_id, album_id, artist_id, name, created_at, updated_at) VALUES %s ", PrepareBatchValuesPG(7, len(songValues)/7))

	_, err := d.MustGetTx().Exec(sql, songValues...)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) CreateThumbnail(thumbnailValues []interface{}) error {
	sql := fmt.Sprintf("INSERT INTO thumbnails (id, entity_type, entity_id, url, width, height, created_at, updated_at) VALUES %s ", PrepareBatchValuesPG(8, len(thumbnailValues)/8))
	_, err := d.MustGetTx().Exec(sql, thumbnailValues...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateAlbum(albumValues []interface{}) error {
	sql := fmt.Sprintf("INSERT INTO albums (id, name, artist_id, spotify_id, created_at, updated_at) VALUES %s ", PrepareBatchValuesPG(6, len(albumValues)/6))
	_, err := d.MustGetTx().Exec(sql, albumValues...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateTopSong(topArtistValues []interface{}) error {
	sql := fmt.Sprintf(`INSERT INTO top_songs (id, user_id, song_id, "order", time_period, created_at, updated_at) VALUES %s `, PrepareBatchValuesPG(7, len(topArtistValues)/7))
	_, err := d.MustGetTx().Exec(sql, topArtistValues...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateRecentlyListened(recentlyListenedValues []interface{}) error {
	sql := fmt.Sprintf(`INSERT INTO recent_listens (id, user_id, created_at, updated_at) VALUES %s `, PrepareBatchValuesPG(4, len(recentlyListenedValues)/4))
	_, err := d.MustGetTx().Exec(sql, recentlyListenedValues...)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) Create(model goutils.Model, values []interface{}) error {
	columnNames := goutils.ColumnNamesExclusive(model)
	tableName := model.TableName()
	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s `, tableName, columnNames, PrepareBatchValuesPG(len(model.TableColumns()), len(values)/len(model.TableColumns())))
	_, err := d.MustGetTx().Exec(sql, values...)
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

func PrepareInStringPG(paramLength int, valueLength int, counter int) string {
	if counter == 0 {
		counter = 1
	}
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
