package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/do4-2022/grobuzin/database"
	"github.com/do4-2022/grobuzin/objectStorage"
	"github.com/do4-2022/grobuzin/routes"
	"github.com/do4-2022/grobuzin/scheduler"

	"github.com/caarlos0/env/v10"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	// rootFsStorageDSN string `env:"ROOT_FS_STORAGE_DSN,notEmpty"`
	LambdoURL			   string `env:"LAMBDO_URL,notEmpty"`
	VMStateURL             string `env:"VM_STATE_URL,notEmpty"`
	FuntionStateStorageDSN string `env:"FUNCTION_STATE_STORAGE_DSN,notEmpty" envDefault:"host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"`
	JWTSecret              string `env:"JWT_SECRET,notEmpty"`
	BuilderEndpoint        string `env:"BUILDER_ENDPOINT,notEmpty"`
	MinioEndpoint          string `env:"MINIO_ENDPOINT,notEmpty"`
	MinioAccessKey         string `env:"MINIO_ACCESS_KEY,notEmpty"`
	MinioSecretKey         string `env:"MINIO_SECRET_KEY,notEmpty"`
	MinioSecure            bool   `env:"MINIO_SECURE" envDefault:"false"`
}

func main() {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("%+v\n", err)
	}

	ctx := context.Background()
	redis := database.InitRedis(cfg.VMStateURL)

	bucketPrefix := "http://"

	if cfg.MinioSecure {
		bucketPrefix = "https://"
	}

	s := &scheduler.Scheduler{
		Redis:   redis,
		Context: &ctx,
		Lambdo: &scheduler.LambdoService{
			URL: cfg.LambdoURL,
			BucketURL: fmt.Sprint(
				bucketPrefix,
				cfg.MinioEndpoint, 
				"/", 
				objectStorage.BucketName,
			),
		},
	}
	//Now inject the scheduler into the routes that need it!

	go func() {
		// every 24 hours we check for
		for {
			time.Sleep(time.Hour * 6)

			log.Println("Ran instance pruning at", time.Now().UTC())
			s.FindAndDestroyUnsused(24)
		}
	}()

	db := database.Init(cfg.FuntionStateStorageDSN)
	r := routes.GetRoutes(db, cfg.JWTSecret, cfg.BuilderEndpoint, getMinioClient(cfg), *s)

	err := r.Run()

	if err != nil {
		panic(err)
	}
}

func getMinioClient(cfg Config) *minio.Client {

	// Initialize minio client object.
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioSecure,
	})
	if err != nil {
		log.Fatalln(err)
	}

	return minioClient
}
