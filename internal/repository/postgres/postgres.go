package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/entity"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/repository"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/timeprovider"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresRepository struct {
	db           *sql.DB
	timeProvider timeprovider.TimeProvider
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
	_, err = tx.Exec("INSERT INTO segments(slug, created_at) VALUES ($1, $2)", slug, p.timeProvider.Now())
	if err != nil {
		return fmt.Errorf("CreateSegment() - tx.Exec(): %w", err)
	}

	// commit changes
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("CreateSegment() - tx.Commit(): %w", err)
	}

	return nil
}

func (p *PostgresRepository) AddSegmentToUsers(slug string, userIDs []int) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("AddSegmentToUsers() - p.db.Begin(): %w", err)
	}

	// check if segment actually exists and get its id and status
	var id int
	var deletedAt sql.NullTime
	row := tx.QueryRow("SELECT id, deleted_at FROM segments WHERE slug=$1", slug)
	if err := row.Scan(&id, &deletedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) { // segment doesn't exist
			return repository.ErrSegmentNotFound
		} else {
			return fmt.Errorf("AddSegmentToUsers() - tx.QueryRow(): %w", err)
		}
	}

	if deletedAt.Valid { // segment is already deleted
		return repository.ErrSegmentAlreadyDeleted
	}

	for _, userID := range userIDs {
		// check if user already has an active segment
		var cnt int
		row = tx.QueryRow(
			`SELECT COUNT(*)
			FROM users_segments
			WHERE user_id=$1
			AND segment_id=(SELECT id FROM segments WHERE slug=$2)
			AND removed_at IS NULL
			AND (expires_at IS NULL OR expires_at > $3)`,
			userID, slug, p.timeProvider.Now(),
		)
		if err := row.Scan(&cnt); err != nil {
			return fmt.Errorf("AddSegmentToUsers() - tx.QueryRow(): %w", err)
		}

		if cnt != 0 { // segment already exists and is active
			continue
		}

		// add segment
		_, err := tx.Exec(
			`INSERT INTO users_segments(segment_id, user_id, added_at)
			VALUES ((SELECT id FROM segments WHERE slug=$1), $2, $3)`,
			slug, userID, p.timeProvider.Now(),
		)

		if err != nil {
			return fmt.Errorf("AddSegmentToUsers() - tx.Exec(): %w", err)
		}
	}

	// commit changes
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("AddSegmentToUsers() - tx.Commit(): %w", err)
	}

	return nil
}

func (p *PostgresRepository) DeleteSegment(slug string) error {
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

		return fmt.Errorf("DeleteSegment() - tx.QueryRow(): %w", err)
	}

	if deletedAt.Valid { // already deleted
		return repository.ErrSegmentAlreadyDeleted
	}

	// mark the segment as deleted
	_, err = tx.Exec("UPDATE segments SET deleted_at=$2 WHERE slug=$1", slug, p.timeProvider.Now())
	if err != nil {
		return fmt.Errorf("DeleteSegment() - tx.Exec(): %w", err)
	}

	// mark active user segments with this segment as removed
	_, err = tx.Exec(
		`UPDATE users_segments SET removed_at=$2, expires_at=NULL
		WHERE segment_id=(SELECT id FROM segments WHERE slug=$1)
		AND removed_at IS NULL
		AND (expires_at IS NULL OR expires_at > $2)`, slug, p.timeProvider.Now())
	if err != nil {
		return fmt.Errorf("DeleteSegment() - tx.Exec(): %w", err)
	}

	// commit changes
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("DeleteSegment() - tx.Commit(): %w", err)
	}

	return nil
}

func (p *PostgresRepository) UpdateUserSegments(userID int, addSegments []entity.SegmentExpiration, removeSegments []entity.SegmentExpiration) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("UpdateUserSegments() - p.db.Begin(): %w", err)
	}
	defer tx.Rollback()

	for _, segment := range addSegments {
		// check segment existence and status and get its id
		var segmentID int
		var deletedAt sql.NullTime
		row := tx.QueryRow("SELECT id, deleted_at FROM segments WHERE slug=$1", segment.Slug)
		if err := row.Scan(&segmentID, &deletedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) { // no such segment at all
				return repository.ErrSegmentNotFound
			}

			return fmt.Errorf("UpdateUserSegments() - tx.QueryRow(): %w", err)
		}

		if deletedAt.Valid { // already deleted
			return repository.ErrSegmentAlreadyDeleted
		}

		// check if user already has the segment
		var cnt int
		row = tx.QueryRow(
			`SELECT COUNT(*)
			FROM users_segments
			WHERE user_id=$1
			AND segment_id=$2
			AND removed_at IS NULL
			AND (expires_at IS NULL OR expires_at > $3)`,
			userID, segmentID, p.timeProvider.Now(),
		)
		if err := row.Scan(&cnt); err != nil {
			return fmt.Errorf("UpdateUserSegments() - tx.QueryRow(): %w", err)
		}

		if cnt != 0 { // segment already exists and is active
			continue
		}

		// add the segment
		var expiresAt sql.NullTime
		if segment.ExpiresAt != nil {
			expiresAt.Time = *segment.ExpiresAt
			expiresAt.Valid = true
		}
		_, err := tx.Exec(
			`INSERT INTO users_segments(segment_id, user_id, added_at, expires_at)
			VALUES ($1, $2, $3, $4)`,
			segmentID, userID, p.timeProvider.Now(), expiresAt,
		)

		if err != nil {
			return fmt.Errorf("UpdateUserSegments() - tx.Exec(): %w", err)
		}
	}

	for _, segment := range removeSegments {
		// check segment existence and status and get its id
		var segmentID int
		var deletedAt sql.NullTime
		row := tx.QueryRow("SELECT id, deleted_at FROM segments WHERE slug=$1", segment.Slug)
		if err := row.Scan(&segmentID, &deletedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) { // no such segment at all
				return repository.ErrSegmentNotFound
			}

			return fmt.Errorf("UpdateUserSegments() - tx.QueryRow(): %w", err)
		}

		if deletedAt.Valid { // already deleted
			return repository.ErrSegmentAlreadyDeleted
		}

		// check if user already has the segment
		var cnt int
		row = tx.QueryRow(
			`SELECT COUNT(*)
			FROM users_segments
			WHERE user_id=$1
			AND segment_id=$2
			AND removed_at IS NULL
			AND (expires_at IS NULL OR expires_at > $3)`,
			userID, segmentID, p.timeProvider.Now(),
		)
		if err := row.Scan(&cnt); err != nil {
			return fmt.Errorf("UpdateUserSegments() - tx.QueryRow(): %w", err)
		}

		if cnt == 0 { // segment doesn't exist
			continue
		}

		// remove the segment
		_, err := tx.Exec(
			`UPDATE users_segments
			SET removed_at=$3
			WHERE user_id=$1
			AND segment_id=$2`,
			userID, segmentID, p.timeProvider.Now(),
		)
		if err != nil {
			return fmt.Errorf("UpdateUserSegments() - tx.Exec(): %w", err)
		}
	}

	// commit changes
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("UpdateUserSegments() - tx.Commit(): %w", err)
	}

	return nil
}

func (p *PostgresRepository) GetActiveUserSegments(userID int) ([]entity.UserSegment, error) {
	rows, err := p.db.Query(
		`SELECT (SELECT slug FROM segments WHERE id=segment_id), added_at, expires_at
		FROM users_segments
		WHERE user_id=$1
		AND removed_at IS NULL
		AND (expires_at IS NULL OR expires_at > $2)`,
		userID, p.timeProvider.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("GetActiveUserSegments() - p.db.Query(): %w", err)
	}

	userSegments := make([]entity.UserSegment, 0, 30)
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

func timeInBounds(t time.Time, timeFrom time.Time, timeTo time.Time) bool {
	return (t.After(timeFrom) || t.Equal(timeFrom)) && t.Before(timeTo)
}

func (p *PostgresRepository) DumpHistory(userID int, timeFrom time.Time, timeTo time.Time) ([]entity.Operation, error) {
	rows, err := p.db.Query(
		`SELECT (SELECT slug FROM segments WHERE id=segment_id), user_id, added_at, removed_at, expires_at
		FROM users_segments
		WHERE user_id=$1`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("DumpHistory() - p.db.Query(): %w", err)
	}

	operations := make([]entity.Operation, 0, 30)
	for rows.Next() {
		var slug string
		var userID int
		var addedAt time.Time
		var removedAt sql.NullTime
		var expiresAt sql.NullTime

		rows.Scan(&slug, &userID, &addedAt, &removedAt, &expiresAt)

		if timeInBounds(addedAt, timeFrom, timeTo) {
			operations = append(operations, entity.Operation{
				UserID:      userID,
				SegmentSlug: slug,
				Type:        entity.AddedOperationType,
				Time:        addedAt,
			})
		}

		if removedAt.Valid && timeInBounds(removedAt.Time, timeFrom, timeTo) {
			operations = append(operations, entity.Operation{
				UserID:      userID,
				SegmentSlug: slug,
				Type:        entity.RemovedOperationType,
				Time:        removedAt.Time,
			})
		}

		if expiresAt.Valid && expiresAt.Time.Before(p.timeProvider.Now()) && timeInBounds(expiresAt.Time, timeFrom, timeTo) {
			operations = append(operations, entity.Operation{
				UserID:      userID,
				SegmentSlug: slug,
				Type:        entity.ExpiredOperationType,
				Time:        expiresAt.Time,
			})
		}
	}

	// sort by operation time ascending
	sort.Slice(operations, func(i, j int) bool { return operations[i].Time.Before(operations[j].Time) })

	return operations, nil
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

	segments := make([]entity.Segment, 0, 30)
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

func New(postgresURL string, timeProvider timeprovider.TimeProvider) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", postgresURL)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &PostgresRepository{
		db:           db,
		timeProvider: timeProvider,
	}, nil
}

func NewWithExistingConnection(db *sql.DB, timeProvider timeprovider.TimeProvider) *PostgresRepository {
	return &PostgresRepository{
		db:           db,
		timeProvider: timeProvider,
	}
}
