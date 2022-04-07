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

func getLobAddressId(recurseId int) (string, error) {
	var lobAddressId string
	if err := postgresClient.QueryRow("SELECT lob_address_id FROM user_info WHERE recurse_id = $1", recurseId).Scan(&lobAddressId); err != nil {
		log.Printf("QueryRow failed: %v\n", err)
		return "", err
	}

	return lobAddressId, nil
}

func getRecurseIds() ([]int, error) {
	var recurseIds []int
	rows, err := postgresClient.Query("SELECT recurse_id FROM user_info")
	if err != nil {
		log.Printf("QueryRow failed: %v\n", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var recurseId int
		err := rows.Scan(&recurseId)
		if err != nil {
			log.Printf("Reading row failed: %v\n", err)
			return nil, err
		}
		recurseIds = append(recurseIds, recurseId)
	}

	err = rows.Err()
	if err != nil {
		log.Printf("Iterating row failed: %v\n", err)
		return nil, err
	}

	return recurseIds, nil
}
