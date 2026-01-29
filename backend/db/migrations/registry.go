package migrations

var registeredMigrations []*Migration

// RegisterMigration registers a migration
func RegisterMigration(m *Migration) {
	registeredMigrations = append(registeredMigrations, m)
}

// GetRegisteredMigrations returns all registered migrations
func GetRegisteredMigrations() []*Migration {
	return registeredMigrations
}
