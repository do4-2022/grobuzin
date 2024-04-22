package objectStorage

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const (
	bucketName     = "functions"
	codeFileSuffix = "/code.json"
	location       = "eu-west-1"
)

type CodeStorageService struct {
	MinioClient *minio.Client
}

func (service *CodeStorageService) Init() {

	ctx := context.Background()
	err := service.MinioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := service.MinioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}
}
func (service *CodeStorageService) PutCode(id uuid.UUID, files map[string]interface{}) {

	contentType := "application/json"
	ctx := context.Background()
	filePath := id.String() + codeFileSuffix

	jsonFiles, err := json.Marshal(files)
	if err != nil {
		log.Fatalln(err)
	}

	reader := bytes.NewReader([]byte(jsonFiles))

	_, err = service.MinioClient.PutObject(ctx, bucketName, filePath, reader, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalln(err)
	}
}

func (service *CodeStorageService) GetCode(id uuid.UUID) (map[string]interface{}, error) {
	ctx := context.Background()
	filePath := id.String() + codeFileSuffix

	object, err := service.MinioClient.GetObject(ctx, bucketName, filePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(object)
	if err != nil {
		return nil, err
	}

	files := make(map[string]interface{})
	err = json.Unmarshal(buf.Bytes(), &files)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (service *CodeStorageService) DeleteCode(id uuid.UUID) error {
	ctx := context.Background()
	filePath := id.String() + codeFileSuffix

	err := service.MinioClient.RemoveObject(ctx, bucketName, filePath, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}
