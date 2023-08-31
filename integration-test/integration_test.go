package integrationtest

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	v1 "github.com/QiZD90/dynamic-customer-segmentation/internal/controller/http/v1"
	"github.com/QiZD90/dynamic-customer-segmentation/internal/entity"
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

var timeBase = time.Date(2000, time.November, 15, 15, 0, 0, 0, time.UTC)
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

	_, err = tx.Exec("DELETE FROM users_segments")
	if err != nil {
		log.Fatal().Msg("purgeDB() - failed to delete from users segments")
	}

	_, err = tx.Exec("DELETE FROM segments")
	if err != nil {
		log.Fatal().Msg("purgeDB() - failed to delete from segments")
	}

	if err := tx.Commit(); err != nil {
		log.Fatal().Msg("purgeDB() - failed to commit transaction")
	}
}

func TestMain(m *testing.M) {
	// Connect to postgres
	var err error
	db, err = sql.Open("pgx", pgURL)
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
	fstorage, err := ondisk.New("WOULD BE CHANGED", "csv/", filestorage.NewUUIDFileStorageNameSupplier())
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

	fstorage.BaseURL = server.URL + "/csv" // dirty hack sorry not sorry

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

func TestCreateSegment(t *testing.T) {
	defer purgeDB(db)

	url := server.URL + "/api/v1/segment/create"
	// First request; should be successfull
	{
		var body bytes.Buffer
		body.WriteString(`{"slug": "AVITO_TEST_SEGMENT"}`)

		r, err := http.Post(url, "application/json", &body)
		assert.NoError(t, err, "TestCreateSegment() - http.Post()")
		defer r.Body.Close()

		expected := v1.JsonStatus{Status: "OK"}
		var got v1.JsonStatus

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestCreateSegment() - failed to unmarshall json")
		}

		assert.Equal(t, http.StatusOK, r.StatusCode)
		assert.Equal(t, expected, got)
	}

	// Second request; should fail with 400 and JsonError
	{
		var body bytes.Buffer
		body.WriteString(`{"slug": "AVITO_TEST_SEGMENT"}`)

		r, err := http.Post(url, "application/json", &body)
		assert.NoError(t, err, "TestCreateSegment() - http.Post()")
		defer r.Body.Close()

		expected := v1.JsonError{StatusCode: http.StatusBadRequest, Message: "Segment already exists"}
		var got v1.JsonError

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestCreateSegment() - failed to unmarshall json")
		}

		assert.Equal(t, http.StatusBadRequest, r.StatusCode)
		assert.Equal(t, expected, got)
	}

	// Third request; should fail with 400 and JsonError
	{
		var body bytes.Buffer
		body.WriteString(`INVALID JSON`)

		r, err := http.Post(url, "application/json", &body)
		assert.NoError(t, err, "TestCreateSegment() - http.Post()")
		defer r.Body.Close()

		expected := v1.JsonError{StatusCode: http.StatusBadRequest, Message: "Error while unmarshalling request JSON"}
		var got v1.JsonError

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestCreateSegment() - failed to unmarshall json")
		}

		assert.Equal(t, http.StatusBadRequest, r.StatusCode)
		assert.Equal(t, expected, got)
	}
}

func TestDeleteSegment(t *testing.T) {
	defer purgeDB(db)

	url := server.URL + "/api/v1/segment/delete"

	// Create segment to be deleted
	assert.NoError(t, s.CreateSegment("AVITO_TEST_SEGMENT"))

	// First request; should be successfull
	{
		var body bytes.Buffer
		body.WriteString(`{"slug": "AVITO_TEST_SEGMENT"}`)

		r, err := http.Post(url, "application/json", &body)
		assert.NoError(t, err, "TestDeleteSegment() - http.Post()")
		defer r.Body.Close()

		expected := v1.JsonStatus{Status: "OK"}
		var got v1.JsonStatus

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestDeleteSegment() - failed to unmarshall json")
		}

		assert.Equal(t, http.StatusOK, r.StatusCode)
		assert.Equal(t, expected, got)
	}

	// Second request; should fail with 400 and JsonError
	{
		var body bytes.Buffer
		body.WriteString(`{"slug": "AVITO_TEST_SEGMENT"}`)

		r, err := http.Post(url, "application/json", &body)
		assert.NoError(t, err, "TestDeleteSegment() - http.Post()")
		defer r.Body.Close()

		expected := v1.JsonError{StatusCode: http.StatusBadRequest, Message: "Segment is already deleted"}
		var got v1.JsonError

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestDeleteSegment() - failed to unmarshall json")
		}

		assert.Equal(t, http.StatusBadRequest, r.StatusCode)
		assert.Equal(t, expected, got)
	}

	// Third request; should fail with 400 and JsonError
	{
		var body bytes.Buffer
		body.WriteString(`INVALID JSON`)

		r, err := http.Post(url, "application/json", &body)
		assert.NoError(t, err, "TestDeleteSegment() - http.Post()")
		defer r.Body.Close()

		expected := v1.JsonError{StatusCode: http.StatusBadRequest, Message: "Error while unmarshalling request JSON"}
		var got v1.JsonError

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestDeleteSegment() - failed to unmarshall json")
		}

		assert.Equal(t, http.StatusBadRequest, r.StatusCode)
		assert.Equal(t, expected, got)
	}

	// Fourth request; should fail with 400 and JsonError
	{
		var body bytes.Buffer
		body.WriteString(`{"slug": "AVITO_SEGMENT_THAT_WAS_NOT_CREATED"}`)

		r, err := http.Post(url, "application/json", &body)
		assert.NoError(t, err, "TestDeleteSegment() - http.Post()")
		defer r.Body.Close()

		expected := v1.JsonError{StatusCode: http.StatusBadRequest, Message: "Segment wasn't found"}
		var got v1.JsonError

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestDeleteSegment() - failed to unmarshall json")
		}

		assert.Equal(t, http.StatusBadRequest, r.StatusCode)
		assert.Equal(t, expected, got)
	}
}

func TestGetSegmentsAndGetActiveSegments(t *testing.T) {
	defer purgeDB(db)

	hourAfterTimeBase := timeBase.Add(time.Hour)
	twoHoursAfterTimeBase := timeBase.Add(2 * time.Hour)
	defer timeProvider.SetTime(timeBase)

	// Create and delete segments
	assert.NoError(t, s.CreateSegment("AVITO_TEST_SEGMENT"))
	assert.NoError(t, s.CreateSegment("AVITO_DELETED_SEGMENT"))
	timeProvider.SetTime(hourAfterTimeBase)
	assert.NoError(t, s.CreateSegment("AVITO_VOICE_MESSAGES"))
	assert.NoError(t, s.DeleteSegment("AVITO_DELETED_SEGMENT"))

	// First request
	{
		r, err := http.Get(server.URL + "/api/v1/segments")
		assert.NoError(t, err, "TestGetSegmentsAndGetActiveSegments() - http.Post()")
		defer r.Body.Close()

		expected := v1.JsonSegments{
			Segments: []entity.Segment{
				{Slug: "AVITO_TEST_SEGMENT", CreatedAt: timeBase, DeletedAt: nil},
				{Slug: "AVITO_VOICE_MESSAGES", CreatedAt: hourAfterTimeBase, DeletedAt: nil},
				{Slug: "AVITO_DELETED_SEGMENT", CreatedAt: timeBase, DeletedAt: &hourAfterTimeBase},
			},
		}
		var got v1.JsonSegments

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestGetSegmentsAndGetActiveSegments() - failed to unmarshall json")
		}

		// sort expected and got by slugs
		sort.Slice(expected.Segments, func(i, j int) bool { return expected.Segments[i].Slug < expected.Segments[j].Slug })
		sort.Slice(got.Segments, func(i, j int) bool { return got.Segments[i].Slug < got.Segments[j].Slug })

		assert.Equal(t, http.StatusOK, r.StatusCode)
		assert.True(t, reflect.DeepEqual(expected, got), "expected: %s, got: %s", expected, got)
	}

	// First request for active segments
	{
		r, err := http.Get(server.URL + "/api/v1/segments/active")
		assert.NoError(t, err, "TestGetSegmentsAndGetActiveSegments() - http.Post()")
		defer r.Body.Close()

		expected := v1.JsonSegments{
			Segments: []entity.Segment{
				{Slug: "AVITO_TEST_SEGMENT", CreatedAt: timeBase, DeletedAt: nil},
				{Slug: "AVITO_VOICE_MESSAGES", CreatedAt: hourAfterTimeBase, DeletedAt: nil},
			},
		}
		var got v1.JsonSegments

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestGetSegmentsAndGetActiveSegments() - failed to unmarshall json")
		}

		// sort expected and got by slugs
		sort.Slice(expected.Segments, func(i, j int) bool { return expected.Segments[i].Slug < expected.Segments[j].Slug })
		sort.Slice(got.Segments, func(i, j int) bool { return got.Segments[i].Slug < got.Segments[j].Slug })

		assert.Equal(t, http.StatusOK, r.StatusCode)
		assert.True(t, reflect.DeepEqual(expected, got), "expected: %s, got: %s", expected, got)
	}

	// Delete all segments
	timeProvider.SetTime(twoHoursAfterTimeBase)
	s.DeleteSegment("AVITO_TEST_SEGMENT")
	s.DeleteSegment("AVITO_VOICE_MESSAGES")

	// Second request
	{
		r, err := http.Get(server.URL + "/api/v1/segments")
		assert.NoError(t, err, "TestGetSegmentsAndGetActiveSegments() - http.Post()")
		defer r.Body.Close()

		expected := v1.JsonSegments{
			Segments: []entity.Segment{
				{Slug: "AVITO_TEST_SEGMENT", CreatedAt: timeBase, DeletedAt: &twoHoursAfterTimeBase},
				{Slug: "AVITO_VOICE_MESSAGES", CreatedAt: hourAfterTimeBase, DeletedAt: &twoHoursAfterTimeBase},
				{Slug: "AVITO_DELETED_SEGMENT", CreatedAt: timeBase, DeletedAt: &hourAfterTimeBase},
			},
		}
		var got v1.JsonSegments

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestGetSegmentsAndGetActiveSegments() - failed to unmarshall json")
		}

		// sort expected and got by slugs
		sort.Slice(expected.Segments, func(i, j int) bool { return expected.Segments[i].Slug < expected.Segments[j].Slug })
		sort.Slice(got.Segments, func(i, j int) bool { return got.Segments[i].Slug < got.Segments[j].Slug })

		assert.Equal(t, http.StatusOK, r.StatusCode)
		assert.True(t, reflect.DeepEqual(expected, got), "expected: %s, got: %s", expected, got)
	}

	// First request for active segments
	{
		r, err := http.Get(server.URL + "/api/v1/segments/active")
		assert.NoError(t, err, "TestGetSegmentsAndGetActiveSegments() - http.Post()")
		defer r.Body.Close()

		expected := v1.JsonSegments{
			Segments: []entity.Segment{},
		}
		var got v1.JsonSegments

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestGetSegmentsAndGetActiveSegments() - failed to unmarshall json")
		}

		// sort expected and got by slugs
		sort.Slice(expected.Segments, func(i, j int) bool { return expected.Segments[i].Slug < expected.Segments[j].Slug })
		sort.Slice(got.Segments, func(i, j int) bool { return got.Segments[i].Slug < got.Segments[j].Slug })

		assert.Equal(t, http.StatusOK, r.StatusCode)
		assert.True(t, reflect.DeepEqual(expected, got), "expected: %s, got: %s", expected, got)
	}
}

func TestCSV(t *testing.T) {
	defer purgeDB(db)
	defer timeProvider.SetTime(timeBase)

	addAndDelete := func(slug string, userID int, timeAdd time.Time, timeDelete time.Time) {
		timeProvider.SetTime(timeAdd)
		assert.NoError(t, s.CreateSegment(slug))
		assert.NoError(t, s.UpdateUserSegments(userID, []entity.SegmentExpiration{{Slug: slug}}, []entity.SegmentExpiration{}))
		timeProvider.SetTime(timeDelete)
		assert.NoError(t, s.DeleteSegment(slug))
	}

	addAndRemove := func(slug string, userID int, timeAdd time.Time, timeDelete time.Time) {
		timeProvider.SetTime(timeAdd)
		assert.NoError(t, s.CreateSegment(slug))
		assert.NoError(t, s.UpdateUserSegments(userID, []entity.SegmentExpiration{{Slug: slug}}, []entity.SegmentExpiration{}))
		timeProvider.SetTime(timeDelete)
		assert.NoError(t, s.UpdateUserSegments(userID, []entity.SegmentExpiration{}, []entity.SegmentExpiration{{Slug: slug}}))
	}

	generateCSVString := func(userID int, operations []entity.Operation) string {
		sb := strings.Builder{}

		for _, o := range operations {
			sb.WriteString(fmt.Sprintf("%d;%s;%s;%s\n", userID, o.SegmentSlug, o.Type, o.Time))
		}

		return sb.String()
	}

	// Create and delete/remove segments
	addAndDelete("AVITO_OUT_OF_BOUNDS", 1021, timeBase.Add(-1*time.Hour*24*30), timeBase.Add(5*time.Hour*24*30))              // -1 month -- +5 months
	addAndRemove("AVITO_IN_BOUNDS", 1021, timeBase, timeBase.Add(3*time.Hour*24*30))                                          // 0 -- +3 months
	addAndDelete("AVITO_ONLY_ADD", 1021, timeBase.Add(2*time.Hour*24*30), timeBase.Add(4*time.Hour*24*30))                    // +2 months -- +4 months
	addAndRemove("AVITO_ONLY_REMOVE", 1021, timeBase.Add(-2*time.Hour*24*30), timeBase.Add(2*time.Hour*24*30+4*time.Hour*24)) // -2 months -- +2 months + 4 days
	addAndRemove("AVITO_TOO_EARLY", 1021, timeBase.Add(-4*time.Hour*24*30), timeBase.Add(-3*time.Hour*24*30))                 // -4 months -- -3 months
	addAndRemove("AVITO_TOO_LATE", 1021, timeBase.Add(7*time.Hour*24*30), timeBase.Add(9*time.Hour*24*30))                    // +7 months -- +9 months

	operations := []entity.Operation{
		{UserID: 1021, SegmentSlug: "AVITO_IN_BOUNDS", Type: entity.AddedOperationType, Time: timeBase},
		{UserID: 1021, SegmentSlug: "AVITO_IN_BOUNDS", Type: entity.RemovedOperationType, Time: timeBase.Add(3 * time.Hour * 24 * 30)},

		{UserID: 1021, SegmentSlug: "AVITO_ONLY_ADD", Type: entity.AddedOperationType, Time: timeBase.Add(2 * time.Hour * 24 * 30)},
		{UserID: 1021, SegmentSlug: "AVITO_ONLY_REMOVE", Type: entity.RemovedOperationType, Time: timeBase.Add(2*time.Hour*24*30 + 4*time.Hour*24)},
	}
	sort.Slice(operations, func(i, j int) bool { return operations[i].Time.Before(operations[j].Time) })
	csv := generateCSVString(1021, operations)

	link := ""
	// Generate CSV
	{
		var b bytes.Buffer
		from := timeBase
		to := timeBase.Add(4 * 30 * time.Hour * 24) // +4 months
		requestJson := fmt.Sprintf(`{
			"user_id": 1021,
			"from": {
				"month": %d,
				"year": %d
			},
			"to": {
				"month": %d,
				"year": %d
			}
		}`, from.Month(), from.Year(), to.Month(), to.Year())
		b.WriteString(requestJson)
		request, err := http.NewRequest("GET", server.URL+"/api/v1/user/csv", &b)
		assert.NoError(t, err, "TestCSV() - http.NewRequest()")

		r, err := http.DefaultClient.Do(request)
		assert.NoError(t, err, "TestCSV() - http.Do()")
		defer r.Body.Close()

		var got v1.JsonLink

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestCSV() - failed to unmarshall json")
		}

		assert.Equal(t, http.StatusOK, r.StatusCode)
		link = got.Link
	}

	// Fetch CSV
	{
		fmt.Println(link)
		r, err := http.Get(link)
		assert.NoError(t, err, "TestCSV() - http.Get()")
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err, "TestCSV() - io.ReadAll()")
		str := string(b)

		assert.Equal(t, http.StatusOK, r.StatusCode)
		assert.Equal(t, csv, str)
	}
}

func TestUserSegmentsAndUpdateUser(t *testing.T) {
	defer purgeDB(db)
	defer timeProvider.SetTime(timeBase)

	yesterday := timeBase.Add(-1 * 24 * time.Hour)
	tomorrow := timeBase.Add(24 * time.Hour)

	userID := 1102
	segments := []struct {
		Slug      string
		CreatedAt time.Time
		AddedAt   time.Time
		DeletedAt *time.Time
		RemovedAt *time.Time
		ExpiresAt *time.Time
	}{
		{Slug: "AVITO_SEGMENT_1", CreatedAt: timeBase.Add(-time.Hour), AddedAt: timeBase.Add(-time.Hour), DeletedAt: nil, ExpiresAt: nil},
		{Slug: "AVITO_SEGMENT_2", CreatedAt: timeBase.Add(-3 * time.Minute), AddedAt: timeBase.Add(-time.Minute), DeletedAt: nil, ExpiresAt: nil},
		{Slug: "AVITO_SEGMENT_3", CreatedAt: timeBase.Add(-time.Hour), AddedAt: timeBase, DeletedAt: nil, ExpiresAt: nil},
		{Slug: "AVITO_EXPIRED", CreatedAt: timeBase.Add(-3 * 24 * time.Hour), AddedAt: timeBase.Add(-2 * 24 * time.Hour), ExpiresAt: &yesterday, DeletedAt: nil},
		{Slug: "AVITO_NOT_YET_EXPIRED", CreatedAt: timeBase.Add(-3 * 24 * time.Hour), AddedAt: timeBase.Add(-2 * 24 * time.Hour), ExpiresAt: &tomorrow, DeletedAt: nil},
		{Slug: "AVITO_DELETED", CreatedAt: timeBase.Add(-3 * 24 * time.Hour), AddedAt: timeBase.Add(-2 * 24 * time.Hour), ExpiresAt: nil, DeletedAt: &yesterday},
		{Slug: "AVITO_REMOVED", CreatedAt: timeBase.Add(-3 * 24 * time.Hour), AddedAt: timeBase.Add(-2 * 24 * time.Hour), ExpiresAt: nil, RemovedAt: &yesterday},
	}

	for _, segment := range segments {
		// Create segment
		{
			timeProvider.SetTime(segment.CreatedAt)

			url := server.URL + "/api/v1/segment/create"

			var body bytes.Buffer
			fmt.Fprintf(&body, `{"slug": "%s"}`, segment.Slug)

			r, err := http.Post(url, "application/json", &body)
			assert.NoError(t, err, "TestUserSegmentsAndUpdateUser() - http.Post()")
			defer r.Body.Close()

			expected := v1.JsonStatus{Status: "OK"}
			var got v1.JsonStatus

			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("TestUserSegmentsAndUpdateUser() - failed to unmarshall json")
			}

			assert.Equal(t, http.StatusOK, r.StatusCode)
			assert.Equal(t, expected, got)
		}

		// Add segment
		{
			timeProvider.SetTime(segment.AddedAt)

			url := server.URL + "/api/v1/user/update"

			request := &v1.JsonUserUpdateRequest{
				UserID: userID,
				AddSegments: []entity.SegmentExpiration{
					{Slug: segment.Slug, ExpiresAt: segment.ExpiresAt},
				},
				RemoveSegments: []entity.SegmentExpiration{},
			}
			var body bytes.Buffer
			if err := json.NewEncoder(&body).Encode(request); err != nil {
				t.Fatalf("TestUserSegmentsAndUpdateUser() - failed to marshall json")
			}

			r, err := http.Post(url, "application/json", &body)
			assert.NoError(t, err, "TestUserSegmentsAndUpdateUser() - http.Post()")
			defer r.Body.Close()

			expected := v1.JsonStatus{Status: "OK"}
			var got v1.JsonStatus

			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("TestUserSegmentsAndUpdateUser() - failed to unmarshall json")
			}

			assert.Equal(t, http.StatusOK, r.StatusCode)
			assert.Equal(t, expected, got)
		}

		// Delete segment
		if segment.DeletedAt != nil {
			timeProvider.SetTime(*segment.DeletedAt)

			url := server.URL + "/api/v1/segment/delete"

			var body bytes.Buffer
			fmt.Fprintf(&body, `{"slug": "%s"}`, segment.Slug)

			r, err := http.Post(url, "application/json", &body)
			assert.NoError(t, err, "TestUserSegmentsAndUpdateUser() - http.Post()")
			defer r.Body.Close()

			expected := v1.JsonStatus{Status: "OK"}
			var got v1.JsonStatus

			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("TestUserSegmentsAndUpdateUser() - failed to unmarshall json")
			}

			assert.Equal(t, http.StatusOK, r.StatusCode)
			assert.Equal(t, expected, got)
		}

		// Remove segment
		if segment.RemovedAt != nil {
			timeProvider.SetTime(*segment.RemovedAt)

			url := server.URL + "/api/v1/user/update"

			request := &v1.JsonUserUpdateRequest{
				UserID:      userID,
				AddSegments: []entity.SegmentExpiration{},
				RemoveSegments: []entity.SegmentExpiration{
					{Slug: segment.Slug},
				},
			}
			var body bytes.Buffer
			if err := json.NewEncoder(&body).Encode(request); err != nil {
				t.Fatalf("TestUserSegmentsAndUpdateUser() - failed to marshall json")
			}

			r, err := http.Post(url, "application/json", &body)
			assert.NoError(t, err, "TestUserSegmentsAndUpdateUser() - http.Post()")
			defer r.Body.Close()

			expected := v1.JsonStatus{Status: "OK"}
			var got v1.JsonStatus

			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("TestUserSegmentsAndUpdateUser() - failed to unmarshall json")
			}

			assert.Equal(t, http.StatusOK, r.StatusCode)
			assert.Equal(t, expected, got)
		}
	}

	// Get user segments
	{
		timeProvider.SetTime(timeBase)

		url := server.URL + "/api/v1/user/segments"

		var body bytes.Buffer
		fmt.Fprintf(&body, `{"user_id": %d}`, userID)

		request, err := http.NewRequest("GET", url, &body)
		assert.NoError(t, err, "TestUserSegmentsAndUpdateUser() - http.NewRequest()")

		r, err := http.DefaultClient.Do(request)
		assert.NoError(t, err, "TestUserSegmentsAndUpdateUser() - http.Do()")
		defer r.Body.Close()

		userSegments := make([]entity.UserSegment, 0, len(segments))
		for _, segment := range segments {
			if segment.DeletedAt != nil || segment.RemovedAt != nil || (segment.ExpiresAt != nil && segment.ExpiresAt.Before(timeProvider.Now())) {
				continue
			}

			userSegments = append(userSegments, entity.UserSegment{
				Slug:      segment.Slug,
				AddedAt:   segment.AddedAt,
				RemovedAt: segment.RemovedAt,
				ExpiresAt: segment.ExpiresAt,
			})
		}

		expected := v1.JsonUserSegments{
			Segments: userSegments,
		}
		var got v1.JsonUserSegments

		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("TestUserSegmentsAndUpdateUser() - failed to unmarshall json")
		}

		// sort by slug ascending
		sort.Slice(expected.Segments, func(i, j int) bool { return expected.Segments[i].Slug < expected.Segments[j].Slug })
		sort.Slice(got.Segments, func(i, j int) bool { return got.Segments[i].Slug < got.Segments[j].Slug })

		assert.Equal(t, http.StatusOK, r.StatusCode)
		assert.True(t, reflect.DeepEqual(expected, got), "expected: %s; got: %s", expected, got)
	}
}
