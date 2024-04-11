package main

import (
	"log"
	"context"

	"github.com/do4-2022/grobuzin/database"
	"github.com/do4-2022/grobuzin/routes"

	"github.com/caarlos0/env/v10"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	// rootFsStorageDSN string `env:"ROOT_FS_STORAGE_DSN,notEmpty"`
	VMStateURL string `env:"VM_STATE_URL,notEmpty"`
	FuntionStateStorageDSN string `env:"FUNCTION_STATE_STORAGE_DSN,notEmpty" envDefault:"host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"`
	JWTSecret              string `env:"JWT_SECRET,notEmpty"`
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

	db := database.Init(cfg.FuntionStateStorageDSN)
	r := routes.GetRoutes(db, cfg.JWTSecret, getMinioClient(cfg))
	redis := database.InitRedis(cfg.VMStateURL)
	redis.Ping(ctx) // TODO define scheduler struct
	

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
