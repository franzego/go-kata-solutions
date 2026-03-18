package main

func main() {
	store := MockStorageService{Timeout: false}
	data := MockMetadataService{Timeout: false}
	auth := MockAuthService{ForceTimeout: false}
	upload := NewUploadService(auth, store, data)
	var file []byte
	err := upload.Upload(file)
	if err != nil {
		// log.Println("Test 1")
		// log.Printf("Raw Error: %v", err)
		// log.Println()
		// log.Println()
		// log.Println("Test 2")

	}

}
