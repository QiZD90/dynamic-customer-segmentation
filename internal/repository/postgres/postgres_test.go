package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/entity"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestCreateSegment(t *testing.T) {
	testCases := []struct {
		name         string
		expectations func(mock sqlmock.Sqlmock)
		slug         string
		expectError  error
	}{
		{
			name: "basic usage",
			expectations: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.
					ExpectQuery(`SELECT COUNT(.+) FROM segments`).
					WithArgs("AVITO_NEW_SEGMENT").
					WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(0))
				mock.
					ExpectExec("INSERT INTO segments").
					WithArgs("AVITO_NEW_SEGMENT").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			slug:        "AVITO_NEW_SEGMENT",
			expectError: nil,
		},

		{
			name: "segment already exists",
			expectations: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.
					ExpectQuery(`SELECT COUNT(.+) FROM segments`).
					WithArgs("AVITO_NEW_SEGMENT").
					WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(1))
				mock.ExpectRollback()
			},
			slug:        "AVITO_NEW_SEGMENT",
			expectError: repository.ErrSegmentAlreadyExists,
		},
	}

	for _, tt := range testCases {
		// Open stub DB connection
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		// Create a mock repository
		repo := &PostgresRepository{db}

		// Build the expectations
		tt.expectations(mock)

		// Execute the method
		err = repo.CreateSegment(tt.slug)
		if err != tt.expectError {
			t.Errorf("wanted error: %s; got error: %s", tt.expectError, err)
		}

		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func TestGetAllSegments(t *testing.T) {
	testCases := []struct {
		name         string
		expectations func(mock sqlmock.Sqlmock)
		expectResult []entity.Segment
		expectError  error
	}{
		{
			name: "basic usage",
			expectations: func(mock sqlmock.Sqlmock) {
				mock.
					ExpectQuery(`SELECT slug, created_at, deleted_at FROM segments`).
					WillReturnRows(sqlmock.
						NewRows([]string{"slug", "created_at", "deleted_at"}).
						AddRow("AVITO_TEST_SEGMENT", time.Time{}, sql.NullTime{}).
						AddRow("AVITO_DELETED_SEGMENT", time.Time{}, sql.NullTime{Valid: true}),
					)
			},
			expectResult: []entity.Segment{
				{Slug: "AVITO_TEST_SEGMENT", CreatedAt: time.Time{}, DeletedAt: nil},
				{Slug: "AVITO_DELETED_SEGMENT", CreatedAt: time.Time{}, DeletedAt: &time.Time{}},
			},
			expectError: nil,
		},
		{
			name: "no rows",
			expectations: func(mock sqlmock.Sqlmock) {
				mock.
					ExpectQuery(`SELECT slug, created_at, deleted_at FROM segments`).
					WillReturnRows(sqlmock.NewRows([]string{"slug", "created_at", "deleted_at"}))
			},
			expectResult: []entity.Segment{},
			expectError:  nil,
		},
	}

	for _, tt := range testCases {
		// Open stub DB connection
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		// Create a mock repository
		repo := &PostgresRepository{db}

		// Build the expectations
		tt.expectations(mock)

		// Execute the method
		segments, err := repo.GetAllSegments()
		if err != tt.expectError {
			t.Errorf("wanted error: %s; got error: %s", tt.expectError, err)
		}

		assert.Equal(t, segments, tt.expectResult)

		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}
