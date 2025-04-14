package store

import "urllite/config/database"

func AutoMigrateTables() {
	migrateUserTable()
	migratePasswordTable()
}

func migrateUserTable() {
	// Create the user table if it doesn't exist
	createUserTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		name TEXT,
		email TEXT,
		mobile TEXT,
		status TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		deleted_at TIMESTAMP
	)`
	if err := database.Session.Query(createUserTable).Exec(); err != nil {
		panic(err)
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
	)`
	if err := database.Session.Query(createPasswordTable).Exec(); err != nil {
		panic(err)
	}
}
