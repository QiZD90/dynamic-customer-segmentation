package postgres

import (
	"database/sql"
	"time"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/entity"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/repository"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresRepository struct {
	db *sql.DB
}

func (p *PostgresRepository) CreateSegment(slug string) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if segment already exists
	row := tx.QueryRow("SELECT COUNT(*) FROM segments WHERE slug=$1", slug)
	var cnt int
	if err := row.Scan(&cnt); err != nil {
		return err
	}

	if cnt >= 1 {
		return repository.ErrAlreadyExists
	}

	// Actually create the segment
	_, err = tx.Exec("INSERT INTO segments(slug) VALUES ($1)", slug)
	if err != nil {
		return err
	}

	// Commit the changes
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *PostgresRepository) CreateSegmentAndEnroll(slug string, userIDs []int) error {
	return nil
}

func (p *PostgresRepository) DeleteSegment(slug string) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if segment already exists
	row := tx.QueryRow("SELECT COUNT(*) FROM segments WHERE slug=$1", slug)
	var cnt int
	if err := row.Scan(&cnt); err != nil {
		return err
	}

	if cnt < 1 {
		return repository.ErrNotFound
	}

	// Mark segment as deleted
	_, err = tx.Exec("UPDATE segments SET deleted_at = NOW() WHERE slug=$1", slug)
	if err != nil {
		return err
	}

	// Mark records with segment as removed
	_, err = tx.Exec("UPDATE users_segments SET removed_at = NOW() WHERE segment_id IN (SELECT id AS segment_if FROM segments WHERE slug=$1)", slug)
	if err != nil {
		return err
	}

	// Commit the changes
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *PostgresRepository) UpdateUserSegments(userID int, addSegments []entity.SegmentExpiration, removeSegments []entity.SegmentExpiration) error {
	return nil
}

func (p *PostgresRepository) GetUserActiveSegments(userID int) ([]entity.Segment, error) {
	rows, err := p.db.Query(
		`SELECT t2.slug, t1.added_at, t1.expires_at
		FROM users_segments AS t1
		JOIN segments AS t2
		ON t2.id = t1.segment_id
		WHERE t1.user_id = $1
		AND t1.removed_at IS NULL
		AND (t1.expires_at IS NULL or t1.expires_at > NOW())`, userID)
	if err != nil {
		return nil, err
	}

	segments := make([]entity.Segment, 0)
	for rows.Next() {
		var segment entity.Segment
		var nullTime sql.NullTime

		rows.Scan(&segment.Slug, &segment.CreatedAt, &nullTime)
		if nullTime.Valid {
			segment.ExpiresAt = &nullTime.Time
		} else {
			segment.ExpiresAt = nil
		}

		segments = append(segments, segment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return segments, nil
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
