package database

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/jackc/pgx"
)

var (
	db        *pgx.ConnPool
	pgxConfig pgx.ConnConfig
)

const (
	schema = "./sql/schema.sql"
)

func init() {
	configFile, _ := os.Open("./config/postgres.json")
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	err := decoder.Decode(&pgxConfig)
	if err != nil {
		log.Println(err)
	}
}

// Connect creates database connection.
func Connect() {
	var err error
	if db, err = pgx.NewConnPool( // creates a new ConnPool. config.ConnConfig is passed through to Connect directly.
		pgx.ConnPoolConfig{
			ConnConfig:     pgxConfig,
			MaxConnections: 8,
		}); err != nil {
		log.Fatalln(err) // Fatalln is equivalent to Println() followed by a call to os.Exit(1)
	}

	if err = ExecSQLScript(schema); err != nil {
		log.Println(err)
	}
	log.Println("SQL Schema was initialized successfully")
}

// Disconnect closes database connection.
func Disconnect() {
	db.Close()
}

// ExecSQLScript execute sql script.
func ExecSQLScript(path string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return err
	}

	if _, err := tx.Exec(string(content)); err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// StartTransaction begins a transation.
func StartTransaction() *pgx.Tx {
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return nil
	}
	return tx
}

// CommitTransaction commits a transation. If commit is not successful, it rollbacks the transaction.
func CommitTransaction(tx *pgx.Tx) {
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		log.Println(err)
	}
}