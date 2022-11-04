package db

import (
	"database/sql"
	"log"
	"regexp"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/traestan/whophone/internal/config"
	"github.com/traestan/whophone/internal/models"
	"go.uber.org/zap"
)

var reAll = regexp.MustCompile(`^((8|\+7|7)[\- ]?)?(\(?\d{3}\)?[\- ]?)`)
var rePrefix = regexp.MustCompile(`^((8|\+7|7)[\- ]?)`)

// Client represents an active db object
type Client struct {
	*sql.DB
	table  string
	logger *zap.Logger
}

// NewClient creates new db instance
func NewClient(cfg *config.DB, logger *zap.Logger) (c *Client, err error) {
	var db *sql.DB
	db, err = sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	//migration
	driver, _ := sqlite3.WithInstance(db, new(sqlite3.Config))
	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"whophone", driver)
	if err != nil {
		logger.Fatal("db connection error", zap.String("error", err.Error()))
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Fatal("db connection error", zap.String("error", err.Error()))
	}
	return &Client{db, cfg.Bucket, logger}, nil
}

// Close closes db connection
func (c *Client) Close() error {
	return c.DB.Close()
}

// GetHash phone search in the database
func (c *Client) GetHash(phoneNumber string) *models.Numbers {
	var code string
	number := new(models.Numbers)

	codeNumber := reAll.FindString(phoneNumber)
	if len(codeNumber) > 3 {
		code = rePrefix.ReplaceAllString(codeNumber, "")
	} else {
		code = codeNumber
	}

	split := reAll.Split(phoneNumber, -1)
	toBetween := split[1]
	qfind := `SELECT operator, region FROM whophone WHERE code= ? and begin<= ? and end >= ?`
	row := c.DB.QueryRow(qfind, code, toBetween, toBetween)
	err := row.Scan(&number.Operator, &number.Region)

	if err != nil {
		if err != sql.ErrNoRows {
			c.logger.Fatal(err.Error())
		}
	}
	return number
}

// PutHash adds the phone to the database
func (c *Client) PutHash(number models.Numbers) error {
	insertPhoneSQL := `INSERT or IGNORE INTO whophone(code, "begin", "end", capacity, operator, region, inn)
	VALUES(?, ?, ?, ?, ?, ?, ?);
	`
	statement, err := c.DB.Prepare(insertPhoneSQL)
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(number.Code, number.Begin, number.End, number.Capacity, number.Operator, number.Region, number.Inn)
	if err != nil {
		log.Fatalln(err.Error())
	}
	return err
}

// StatsInfo information on added records after updating the database
func (c *Client) StatsInfo() error {
	var countRows int
	qCount := `SELECT count(id) as cnt FROM whophone`
	row := c.DB.QueryRow(qCount)

	err := row.Scan(&countRows)
	if err != nil {
		c.logger.Fatal(err.Error())
	}
	c.logger.Info("Number of rows in the table ", zap.Int("number", countRows))
	return nil
}

// ClearOld clearing the database of old numbers
func (c *Client) ClearOld() (bool, error) {
	deleteAll := `DELETE FROM whophone`
	_, err := c.DB.Exec(deleteAll)
	if err != nil {
		log.Fatalln(err.Error())
		return false, err
	}
	return true, nil
}
