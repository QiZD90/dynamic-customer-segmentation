package ondisk

import (
	"net/url"
	"os"
	"path"
	"time"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/filestorage"
)

type OnDiskFileStorage struct {
	BaseURL       string
	DirectoryPath string
	NameSupplier  filestorage.FileStorageNameSupplier
}

func (f *OnDiskFileStorage) StoreCSV(csv string, userID int, timeFrom time.Time, timeTo time.Time) (string, error) {
	filename := f.NameSupplier.GenerateFileName(userID, timeFrom, timeTo)
	path := path.Join(f.DirectoryPath, filename)

	if err := os.WriteFile(path, []byte(csv), 0777); err != nil {
		return "", err
	}

	csvURL, err := url.JoinPath(f.BaseURL, filename)
	if err != nil {
		return "", err
	}

	return csvURL, nil
}

func New(baseURL string, directoryPath string, nameSupplier filestorage.FileStorageNameSupplier) (*OnDiskFileStorage, error) {
	err := os.MkdirAll(directoryPath, 0777)
	if err != nil {
		return nil, err
	}

	return &OnDiskFileStorage{
		BaseURL:       baseURL,
		DirectoryPath: directoryPath,
		NameSupplier:  nameSupplier,
	}, nil
}
