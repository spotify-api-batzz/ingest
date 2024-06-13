package database

import (
	"context"
	"fmt"
	"spotify/models"
	"spotify/utils"

	"github.com/batzz-00/goutils/logger"

	"github.com/jackc/pgx/v4"
	stdlib "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

type DatabaseAuth struct {
	User     string
	IP       string
	Password string
	Port     string
	Table    string
}

type Database struct {
	DB   *sqlx.DB
	Tx   *sqlx.Tx
	Auth DatabaseAuth
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
func (db *Database) Connect() error {

	url := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s", db.Auth.IP, db.Auth.Port, db.Auth.Table, db.Auth.User, db.Auth.Password)

	connConf, err := pgx.ParseConfig(url)
	if err != nil {
		return err
	}
	connConf.PreferSimpleProtocol = true

	nativeDB := stdlib.OpenDB(*connConf)
	sqlxDb := sqlx.NewDb(nativeDB, "pgx")
	err = sqlxDb.Ping()
	if err != nil {
		logger.Log(err, logger.Error)
		// do something here
		return err
	}

	db.DB = sqlxDb
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
	sql := fmt.Sprintf("SELECT * FROM users WHERE spotify_id IN (%s)", utils.PrepareInStringPG(1, len(names), 1))
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
	sql := fmt.Sprintf("SELECT * FROM songs WHERE spotify_id IN (%s)", utils.PrepareBatchValuesPG(1, len(spotifyIDs)))
	err := d.MustGetTx().Select(&songs, sql, spotifyIDs...)
	if err != nil {
		return nil, err
	}
	return songs, nil
}

func (d *Database) FetchAlbumsBySpotifyID(spotifyIDs []interface{}) ([]models.Album, error) {
	albums := []models.Album{}
	sql := fmt.Sprintf("SELECT * FROM albums WHERE spotify_id IN (%s)", utils.PrepareBatchValuesPG(1, len(spotifyIDs)))
	err := d.MustGetTx().Select(&albums, sql, spotifyIDs...)
	if err != nil {
		return nil, err
	}
	return albums, nil
}

func (d *Database) FetchArtistsBySpotifyID(spotifyIDs []interface{}) ([]models.Artist, error) {
	artists := []models.Artist{}
	sql := fmt.Sprintf("SELECT * FROM artists WHERE spotify_id IN (%s)", utils.PrepareBatchValuesPG(1, len(spotifyIDs)))
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
	columnNames := utils.ColumnNamesExclusive(&models.RecentListen{})
	tableName := (&models.RecentListen{}).TableName()
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 AND played_at >= $2 AND played_at IN (%s)", columnNames, tableName, utils.PrepareInStringPG(1, len(recentListenedToIDs), 3))
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
	sql := fmt.Sprintf("SELECT * FROM thumbnails WHERE entity_id IN (%s)", utils.PrepareInStringPG(1, len(entityIDs), 1))
	err := d.MustGetTx().Select(&thumbnails, sql, entityIDs...)
	if err != nil {
		return nil, err
	}
	return thumbnails, nil
}

func (d *Database) Create(model models.Model, values []interface{}) error {
	columnNames := utils.ColumnNamesExclusive(model)
	tableName := model.TableName()
	colLength := len(utils.ReflectColumns(model))

	preppedValues := utils.PrepareBatchValuesPG(colLength, len(values)/colLength)
	fmt.Println(preppedValues)
	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s `, tableName, columnNames, preppedValues)
	_, err := d.MustGetTx().Exec(sql, values...)
	if err != nil {
		return err
	}
	return nil
}
