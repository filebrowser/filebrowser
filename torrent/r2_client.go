package torrent

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	client *s3.Client
	once   sync.Once
)

// InitializeClient initializes the S3 client and ensures it's only done once.
func InitializeClient(accountId, accessKeyId, accessKeySecret string) {
	once.Do(func() {
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
			config.WithRegion("auto"),
		)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}

		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
		})
	})
}

// GetClient returns the initialized S3 client.
func GetClient() *s3.Client {
	if client == nil {
		log.Fatal("S3 client is not initialized. Call InitializeClient first.")
	}
	return client
}

// ListFiles lists all files in the specified bucket.
func listFiles(bucketName string) ([]types.Object, error) {
	client := GetClient()
	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	return output.Contents, nil
}

// UploadFile uploads a file to the specified bucket.
func uploadHandler(bucketName, key, filePath string, c chan string) error { //nolint:interfacer
	client := GetClient()

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Upload the file
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &key,
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}
