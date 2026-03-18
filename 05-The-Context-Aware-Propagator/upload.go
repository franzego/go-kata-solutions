package main

import "fmt"

type UploadService struct {
	Auth  MockAuthService
	Store MockStorageService
	Data  MockMetadataService
}

func NewUploadService(auth MockAuthService, store MockStorageService, data MockMetadataService) UploadService {
	return UploadService{
		Auth:  auth,
		Store: store,
		Data:  data,
	}
}

func (u *UploadService) Upload(file []byte) error {
	err := u.Auth.Authenticate()
	if err != nil {
		return fmt.Errorf("upload failed at the authentication layer: %w", err)
	}
	err = u.Data.Metadata()
	if err != nil {
		return fmt.Errorf("upload failed at the metadata layer: %w", err)
	}
	err = u.Store.StorageService()
	if err != nil {
		return fmt.Errorf("upload failed at the storage layer: %w", err)
	}
	return nil
}
