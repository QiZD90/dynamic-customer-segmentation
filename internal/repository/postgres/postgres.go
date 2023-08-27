package postgres

import (
	"database/sql"
	"errors"
	"fmt"
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
		return fmt.Errorf("CreateSegment() - p.db.Begin(): %w", err)
	}
	defer tx.Rollback()

	// check if there is a segment under this slug
	var cnt int
	row := tx.QueryRow("SELECT COUNT(*) FROM segments WHERE slug=$1", slug)
	if err := row.Scan(&cnt); err != nil {
		return fmt.Errorf("CreateSegment() - tx.QueryRow(): %w", err)
	}

	if cnt != 0 {
		return repository.ErrSegmentAlreadyExists
	}

	// create the segment
	_, err = tx.Exec("INSERT INTO segments(slug) VALUES ($1)", slug)
	if err != nil {
		return fmt.Errorf("CreateSegment() - tx.Exec(): %w", err)
	}

	// commit changes
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("CreateSegment() - tx.Commit(): %w", err)
	}

	return nil
}

func (p *PostgresRepository) AddSegmentToUsers(slug string, userIDs []int) error { // TODO:
	return nil
}

func (p *PostgresRepository) DeleteSegment(slug string) error { // TODO:
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("DeleteSegment() - p.db.Begin(): %w", err)
	}
	defer tx.Rollback()

	// get the deletion time of this segment to check its status
	var deletedAt sql.NullTime
	row := tx.QueryRow("SELECT deleted_at FROM segments WHERE slug=$1", slug)
	if err := row.Scan(&deletedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) { // no such segment at all
			return repository.ErrSegmentNotFound
		}

		return fmt.Errorf("CreateSegment() - tx.QueryRow(): %w", err)
	}

	if deletedAt.Valid { // already deleted
		return repository.ErrSegmentAlreadyDeleted
	}

	// mark the segment as deleted
	_, err = tx.Exec("UPDATE segments SET deleted_at=NOW() WHERE slug=$1", slug)
	if err != nil {
		return fmt.Errorf("DeleteSegment() - tx.Exec(): %w", err)
	}

	// mark active user segments with this segment as removed
	_, err = tx.Exec(
		`UPDATE users_segments SET removed_at=NOW(), expires_at=NULL
		WHERE segment_id=(SELECT id FROM segments WHERE slug=$1)
		AND removed_at IS NULL
		AND (expires_at IS NULL OR expires_at > NOW())`, slug)
	if err != nil {
		return fmt.Errorf("DeleteSegment() - tx.Exec(): %w", err)
	}

	// commit changes
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("DeleteSegment() - tx.Commit(): %w", err)
	}

	return nil
}

func (p *PostgresRepository) UpdateUserSegments(userID int, addSegments []entity.SegmentExpiration, removeSegments []entity.SegmentExpiration) error { // TODO:
	return nil
}

func (p *PostgresRepository) GetActiveUserSegments(userID int) ([]entity.UserSegment, error) {
	rows, err := p.db.Query(
		`SELECT (SELECT slug FROM segments WHERE id=segment_id), added_at, expires_at
		FROM users_segments
		WHERE user_id=$1
		AND removed_at IS NULL
		AND (expires_at IS NULL OR expires_at > NOW())`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("GetActiveUserSegments() - p.db.Query(): %w", err)
	}

	userSegments := make([]entity.UserSegment, 0)
	for rows.Next() {
		var userSegment entity.UserSegment
		var expiresAt sql.NullTime
		rows.Scan(&userSegment.Slug, &userSegment.AddedAt, &expiresAt)

		if expiresAt.Valid {
			userSegment.ExpiresAt = &expiresAt.Time
		} else {
			userSegment.ExpiresAt = nil
		}

		userSegments = append(userSegments, userSegment)
	}

	return userSegments, nil
}

func (p *PostgresRepository) DumpHistory(userIDs []int, timeFrom time.Time, timeTo time.Time) ([]entity.Operation, error) { // TODO
	return nil, nil
}

func (p *PostgresRepository) GetAllActiveSegments() ([]entity.Segment, error) {
	rows, err := p.db.Query("SELECT slug, created_at FROM segments WHERE deleted_at IS NULL")
	if err != nil {
		return nil, fmt.Errorf("GetAllActiveSegments() - p.db.Query(): %w", err)
	}

	segments := make([]entity.Segment, 0)
	for rows.Next() {
		var segment entity.Segment
		rows.Scan(&segment.Slug, &segment.CreatedAt)

		segments = append(segments, segment)
	}

	return segments, nil
}

func (p *PostgresRepository) GetAllSegments() ([]entity.Segment, error) {
	rows, err := p.db.Query("SELECT slug, created_at, deleted_at FROM segments")
	if err != nil {
		return nil, fmt.Errorf("GetAllSegments() - p.db.Query(): %w", err)
	}

	segments := make([]entity.Segment, 0)
	for rows.Next() {
		var segment entity.Segment
		var deletedAt sql.NullTime
		rows.Scan(&segment.Slug, &segment.CreatedAt, &deletedAt)

		if deletedAt.Valid {
			segment.DeletedAt = &deletedAt.Time
		} else {
			segment.DeletedAt = nil
		}

		segments = append(segments, segment)
	}

	return segments, nil
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
