package store

import "urllite/config/database"

func AutoMigrateTables() {
	migrateUserTable()
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
