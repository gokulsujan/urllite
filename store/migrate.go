package store

import (
	"log"
	"urllite/config/database"
)

func AutoMigrateTables() {
	migrateUserTable()
	migratePasswordTable()
	migrateUrlTable()
	migrateUrlLogTable()
	migrateOtpTable()
}

func migrateUserTable() {
	// Create the user table if it doesn't exist
	createUserTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		name TEXT,
		email TEXT,
		verified_email TEXT,
		mobile TEXT,
		status TEXT,
		role TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		deleted_at TIMESTAMP
	);`

	session, err := database.CreateSession()
	if err != nil {
		log.Fatal("Unable to create session:", err.Error())
	}
	defer session.Close()
	if err := session.Query(createUserTable).Exec(); err != nil {
		log.Fatal("Unable to create user table:", err.Error())
	}
}

func migratePasswordTable() {
	// Create the password table if it doesn't exist
	createPasswordTable := `
	CREATE TABLE IF NOT EXISTS passwords (
		id UUID PRIMARY KEY,
		user_id UUID,
		hashed_password TEXT,
		status TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		deleted_at TIMESTAMP
	);`

	session, err := database.CreateSession()
	if err != nil {
		log.Fatal("Unable to create session:", err.Error())
	}
	defer session.Close()
	if err := session.Query(createPasswordTable).Exec(); err != nil {
		log.Fatal("Unable to create password table:", err.Error())
	}
}

func migrateUrlTable() {
	// Create the url table if it doesn't exist
	createUrlTable := `
	CREATE TABLE IF NOT EXISTS urls (
		id UUID PRIMARY KEY,
		user_id UUID,
		long_url TEXT,
		short_url TEXT,
		status TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		deleted_at TIMESTAMP
	);`

	session, err := database.CreateSession()
	if err != nil {
		log.Fatal("Unable to create session:", err.Error())
	}
	defer session.Close()
	if err := session.Query(createUrlTable).Exec(); err != nil {
		log.Fatal("Unable to create url table:", err.Error())
	}
}

func migrateUrlLogTable() {
	// Create the url table if it doesn't exist
	createUrlLogTable := `
	CREATE TABLE IF NOT EXISTS url_logs (
	url_id UUID,
	id UUID,
	visited_at TIMESTAMP,
	redirect_status TEXT,
	http_status_code INT,
	client_ip TEXT,
	city TEXT,
	country TEXT,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	deleted_at TIMESTAMP,
	PRIMARY KEY ((url_id), created_at, id)
	) WITH CLUSTERING ORDER BY (created_at DESC);`

	session, err := database.CreateSession()
	if err != nil {
		log.Fatal("Unable to create session:", err.Error())
	}
	defer session.Close()
	if err := session.Query(createUrlLogTable).Exec(); err != nil {
		log.Fatal("Unable to create url table:", err.Error())
	}
}

func migrateOtpTable() {
	createUrlLogTable := `
	CREATE TABLE IF NOT EXISTS otp (
		id UUID PRIMARY KEY,
		user_id UUID,
		key TEXT,
		otp TEXT,
		status TEXT,
		created_at TIMESTAMP,
		expired_at TIMESTAMP
	);`

	session, err := database.CreateSession()
	if err != nil {
		log.Fatal("Unable to create session:", err.Error())
	}
	defer session.Close()
	if err := session.Query(createUrlLogTable).Exec(); err != nil {
		log.Fatal("Unable to create url table:", err.Error())
	}
}
