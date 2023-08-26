package postgres

import (
	"database/sql"
	"time"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/entity"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresRepository struct {
	db *sql.DB
}

func (p *PostgresRepository) CreateSegment(slug string) error {
	return nil
}

func (p *PostgresRepository) CreateSegmentAndEnroll(slug string, userIDs []int) error {
	return nil
}

func (p *PostgresRepository) DeleteSegment(slug string) error {
	return nil
}

func (p *PostgresRepository) UpdateUserSegments(userID int, addSegments []entity.SegmentExpiration, removeSegments []entity.SegmentExpiration) error {
	return nil
}

func (p *PostgresRepository) GetUserActiveSegments(userID int) ([]entity.Segment, error) {
	return nil, nil
}

func (p *PostgresRepository) DumpHistory(userIDs []int, timeFrom time.Time, timeTo time.Time) ([]entity.Operation, error) {
	return nil, nil
}

func New(postgresURL string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", postgresURL)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &PostgresRepository{
		db: db,
	}, nil
}
