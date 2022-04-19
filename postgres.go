package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const recurseCenterRecurseId = 0

type PostgresClient struct {
}

var postgresClient = &PostgresClient{}

var db *sql.DB

func (*PostgresClient) setupPostgresConnection() error {
	var err error
	db, err = sql.Open("pgx", os.Getenv("PG_DATABASE_URL"))
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		return err
	}

	db.SetConnMaxLifetime(time.Minute)

	if err = db.Ping(); err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS user_info (recurse_id int UNIQUE NOT NULL, lob_address_id text UNIQUE NOT NULL, user_name text NOT NULL, user_email text NOT NULL)")
	return err
}

func (*PostgresClient) getUserInfo(recurseId int) (lobAddressId string, acceptsPhysicalMail bool, numCredits int, err error) {
	if err = db.QueryRow("SELECT lob_address_id, accepts_physical_mail, num_credits FROM user_info WHERE recurse_id = $1", recurseId).Scan(&lobAddressId, &acceptsPhysicalMail, &numCredits); err != nil {
		log.Printf("QueryRow failed: %v\n", err)
		return "", false, 0, err
	}

	return
}

func (*PostgresClient) getLobAddressId(recurseId int) (string, error) {
	var lobAddressId string
	if err := db.QueryRow("SELECT lob_address_id FROM user_info WHERE recurse_id = $1", recurseId).Scan(&lobAddressId); err != nil {
		log.Printf("QueryRow failed: %v\n", err)
		return "", err
	}

	return lobAddressId, nil
}

func (*PostgresClient) getCredits(recurseId int) (int, error) {
	var credits int
	if err := db.QueryRow("SELECT num_credits FROM user_info WHERE recurse_id=$1", recurseId).Scan(&credits); err != nil {
		log.Printf("QueryRow failed: %v\n", err)
		return -1, err
	}

	return credits, nil
}

func (*PostgresClient) decrementCredits(recurseId int) error {
	if _, err := db.Exec(
		"UPDATE user_info SET num_credits = num_credits - 1 WHERE recurse_id = $1",
		recurseId); err != nil {
		return err
	}
	return nil
}

func (*PostgresClient) getContacts() ([]*Contact, error) {

	var contacts []*Contact
	rows, err := db.Query("SELECT recurse_id, accepts_physical_mail, user_name, user_email FROM user_info")
	if err != nil {
		log.Printf("QueryRow failed: %v\n", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		contact := new(Contact)
		err := rows.Scan(&contact.RecurseId, &contact.AcceptsPhysicalMail, &contact.Name, &contact.Email)
		if err != nil {
			log.Printf("Reading row failed: %v\n", err)
			return nil, err
		}
		contacts = append(contacts, contact)
	}

	err = rows.Err()
	if err != nil {
		log.Printf("Iterating row failed: %v\n", err)
		return nil, err
	}

	return contacts, nil
}

func (*PostgresClient) insertUser(recurseId int, lobAddressId, userName, userEmail string, acceptsPhysicalMail bool) error {
	if _, err := db.Exec(
		"INSERT INTO user_info (recurse_id, lob_address_id, accepts_physical_mail, user_name, user_email) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (recurse_id) DO UPDATE SET lob_address_id = excluded.lob_address_id, accepts_physical_mail = excluded.accepts_physical_mail, user_name = excluded.user_name, user_email = excluded.user_email",
		recurseId,
		lobAddressId,
		acceptsPhysicalMail,
		userName,
		userEmail); err != nil {
		return err
	}
	return nil
}

func (*PostgresClient) deleteUser(recurseId int) error {
	if _, err := db.Exec(
		"DELETE FROM user_info WHERE recurse_id = $1",
		recurseId); err != nil {
		return err
	}
	return nil
}
