package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

const (
	defaultHost     = "localhost"
	defaultPort     = 5432
	defaultUser     = "postgres"
	defaultPassword = "password"
	defaultDBName   = "reservaciones_vuelos"
)

var once sync.Once
var dbInstance *sql.DB

// LoadEnv loads environment variables from a .env file
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using default values")
	}
}

// CreateDatabaseConnection establishes and returns a singleton DB connection
func CreateDatabaseConnection() (*sql.DB, error) {
	var err error
	once.Do(func() {
		LoadEnv()

		host := getEnv("DB_HOST", defaultHost)
		port := getEnv("DB_PORT", fmt.Sprintf("%d", defaultPort))
		user := getEnv("DB_USER", defaultUser)
		password := getEnv("DB_PASSWORD", defaultPassword)
		dbname := getEnv("DB_NAME", defaultDBName)

		psqlInfo := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)

		dbInstance, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			err = fmt.Errorf("unable to open DB: %v", err)
			return
		}

		err = dbInstance.Ping()
		if err != nil {
			err = fmt.Errorf("unable to ping DB: %v", err)
			return
		}

		log.Println("Database connection established")
	})

	return dbInstance, err
}

// SetIsolationLevel sets the desired isolation level for the database connection
func SetIsolationLevel(db *sql.DB, level string) error {
	query := fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", level)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to set isolation level: %v", err)
	}
	log.Printf("Isolation level set to %s", level)
	return nil
}

// getEnv retrieves environment variables or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
