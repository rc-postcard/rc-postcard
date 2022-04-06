package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var postgresClient *sql.DB

func setupPostgresConnection() error {
	var err error
	postgresClient, err = sql.Open("pgx", os.Getenv("PG_DATABASE_URL"))
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		return err
	}

	if err = postgresClient.Ping(); err != nil {
		return err
	}

	_, err = postgresClient.Exec("CREATE TABLE IF NOT EXISTS user_info (recurse_id int UNIQUE NOT NULL, lob_address_id text UNIQUE NOT NULL)")
	return err
}
