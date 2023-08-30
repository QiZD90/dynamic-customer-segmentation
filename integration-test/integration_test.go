package integrationtest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	v1 "github.com/QiZD90/dynamic-customer-segmentation/internal/controller/http/v1"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/filestorage"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/filestorage/ondisk"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/repository/postgres"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/service"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/timeprovider/fixedtimeprovider"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/userservice/usermicroservice"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

const pgURL = "postgresql://testuser:testuserpassword@test-postgres:5432/testdb?sslmode=disable"

var timeBase = time.Date(2023, time.November, 15, 15, 0, 0, 0, time.UTC)
var timeProvider = fixedtimeprovider.New(timeBase)

var db *sql.DB
var s service.Service
var server *httptest.Server

func purgeDB(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal().Msg("purgeDB() - failed to create transaction")
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM segments")
	if err != nil {
		log.Fatal().Msg("purgeDB() - failed to delete from segments")
	}

	_, err = tx.Exec("DELETE FROM users_segments")
	if err != nil {
		log.Fatal().Msg("purgeDB() - failed to delete from users segments")
	}

	if err := tx.Commit(); err != nil {
		log.Fatal().Msg("purgeDB() - failed to commit transaction")
	}
}

func TestMain(m *testing.M) {
	// Connect to postgres
	db, err := sql.Open("pgx", pgURL)
	if err != nil {
		log.Fatal().Msg("Failed to open connection to postgres test instance")
	}

	err = db.Ping()
	if err != nil {
		log.Fatal().Msg("Postgres test instance ping failed")
	}

	// Migrate up to date
	log.Info().Msg("Starting migrations...")
	migration, err := migrate.New("file://../migrations", pgURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load migrations")
	}

	if err := migration.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info().Msg("Already up to date")
		} else {
			log.Fatal().Err(err).Msg("")
		}
	}

	if src_err, db_err := migration.Close(); src_err != nil || db_err != nil {
		log.Fatal().AnErr("src_err", src_err).AnErr("db_err", db_err).Msg("")
	}

	// Create repo
	repo := postgres.NewWithExistingConnection(db, timeProvider)

	// Create fstorage
	fstorage, err := ondisk.New("http://localhost:80/csv/", "csv/", filestorage.NewUUIDFileStorageNameSupplier())
	if err != nil {
		log.Fatal().Msg("Failed to create ondisk file storage")
	}

	// Connect to user service
	userService, err := usermicroservice.New("http://usermicroservice:80/")
	if err != nil {
		log.Fatal().Msg("Failed to connect to usermicroservice")
	}

	// Create the service
	s = service.New(repo, fstorage, userService)

	// Create the mux and start the server
	mux := v1.NewMux(s)
	server = httptest.NewServer(mux)

	fmt.Println("Listening at " + server.URL)

	// Run the tests
	code := m.Run()

	server.Close()

	os.Exit(code)
}

func TestHealth(t *testing.T) {
	r, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("TestHealth() - unexpected error: %s", err)
	}
	defer r.Body.Close()

	expected := v1.JsonStatus{Status: "OK"}
	var got v1.JsonStatus

	if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
		t.Fatalf("TestHealth() - failed to unmarshall json")
	}

	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.Equal(t, expected, got)
}
