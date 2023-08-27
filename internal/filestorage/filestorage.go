package filestorage

import "time"

type FileStorage interface {
	// StoreCSV stores supplied CSV in string format and returns the URL of the resource
	StoreCSV(csv string, userID int, timeFrom time.Time, timeTo time.Time) (string, error)
}
