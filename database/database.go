package database

import (
	"context"

	"github.com/jinzhu/gorm"
)

// Open creates a new connection with the given config
func Open(config Config) (*gorm.DB, error) {
	db, err := gorm.Open(config.Driver, config.ConnectionString)
	if err != nil {
		return nil, err
	}

	db.LogMode(config.Log)

	// Plural table names are lame
	db.SingularTable(true)

	// Do not allow update or delete to be called without a where clause.
	db.BlockGlobalUpdate(true)

	return db, nil
}

// Transact will execute the given function within a database transaction,
// and handle commits or rollbacks as necessary
func Transact(ctx context.Context, db *gorm.DB, f func(ctx context.Context, tx *gorm.DB) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := f(ctx, tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
