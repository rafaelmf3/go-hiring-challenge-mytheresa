package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TestDB struct {
	DB    *gorm.DB
	Close func() error
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupTestDB creates a test database with migrations
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	dbName := getEnvOrDefault("POSTGRES_DB", "challenge")
	user := getEnvOrDefault("POSTGRES_USER", "postgres")
	password := getEnvOrDefault("POSTGRES_PASSWORD", "password")
	port := getEnvOrDefault("POSTGRES_PORT", "5432")

	db, closeDB := database.New(user, password, dbName, port)

	if err := dropTables(db); err != nil {
		closeDB()
		t.Fatalf("Failed to drop tables: %v", err)
	}

	if err := runMigrations(db); err != nil {
		closeDB()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return &TestDB{
		DB:    db,
		Close: closeDB,
	}
}

func dropTables(db *gorm.DB) error {
	ctx := context.Background()

	tables := []string{"product_variants", "products", "categories"}
	for _, table := range tables {
		if err := db.WithContext(ctx).Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}
	return nil
}

// CleanupDatabase truncates all tables for a clean test state
func (tdb *TestDB) CleanupDatabase(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	tables := []string{"product_variants", "products", "categories"}
	for _, table := range tables {
		if err := tdb.DB.WithContext(ctx).Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			t.Logf("Warning: failed to truncate %s: %v", table, err)
		}
	}
}

// SeedTestData seeds the database with test data
func (tdb *TestDB) SeedTestData(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	categories := []models.Category{
		{Code: "CLOTHING", Name: "Clothing"},
		{Code: "SHOES", Name: "Shoes"},
		{Code: "ACCESSORIES", Name: "Accessories"},
	}

	for _, cat := range categories {
		if err := tdb.DB.WithContext(ctx).Create(&cat).Error; err != nil {
			t.Fatalf("Failed to seed category: %v", err)
		}
	}

	var clothing, shoes, accessories models.Category
	tdb.DB.WithContext(ctx).Where("code = ?", "CLOTHING").First(&clothing)
	tdb.DB.WithContext(ctx).Where("code = ?", "SHOES").First(&shoes)
	tdb.DB.WithContext(ctx).Where("code = ?", "ACCESSORIES").First(&accessories)

	products := []models.Product{
		{Code: "PROD001", Price: decimal.NewFromFloat(10.99), CategoryID: &clothing.ID},
		{Code: "PROD002", Price: decimal.NewFromFloat(12.49), CategoryID: &shoes.ID},
		{Code: "PROD003", Price: decimal.NewFromFloat(8.99), CategoryID: &accessories.ID},
	}

	for _, prod := range products {
		if err := tdb.DB.WithContext(ctx).Create(&prod).Error; err != nil {
			t.Fatalf("Failed to seed product: %v", err)
		}
	}
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Category{},
		&models.Product{},
		&models.Variant{},
	)
}
