package filestorage

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

type FileStorageNameSupplier interface {
	GenerateFileName(userID int, timeFrom time.Time, timeTo time.Time) string
}

type UUIDFileStorageNameSupplier struct {
}

func (u *UUIDFileStorageNameSupplier) GenerateFileName(userID int, timeFrom time.Time, timeTo time.Time) string {
	// !!NOTE!!: it panics on error but it's intentional because something has to
	// go VERY wrong for it to fail
	return uuid.Must(uuid.NewV4()).String() + ".csv"
}

func NewUUIDFileStorageNameSupplier() *UUIDFileStorageNameSupplier {
	return &UUIDFileStorageNameSupplier{}
}

type TextFormatNameSupplier struct {
}

func (u *TextFormatNameSupplier) GenerateFileName(userID int, timeFrom time.Time, timeTo time.Time) string {
	return fmt.Sprintf("%d--%d.%d-%d.%d.csv", userID, timeFrom.Month(), timeFrom.Year(), timeTo.Month(), timeTo.Year())
}

func NewTextFormatNameSupplier() *TextFormatNameSupplier {
	return &TextFormatNameSupplier{}
}
