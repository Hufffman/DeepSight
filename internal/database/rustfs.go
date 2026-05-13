package database

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	"DeepSight/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var rustFS *RustFS

type RustFS struct {
	client *s3.Client
	bucket string
	config config.RustFSConfig
}

func InitializeRustFS(cfg *config.RustFSConfig) error {
	if cfg == nil {
		return fmt.Errorf("rustfs config is nil")
	}
	if cfg.Endpoint == "" {
		return fmt.Errorf("rustfs endpoint is required")
	}
	if cfg.AccessKeyID == "" {
		return fmt.Errorf("rustfs access key is required")
	}
	if cfg.SecretAccessKey == "" {
		return fmt.Errorf("rustfs secret access key is required")
	}

	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	endpoint := cfg.Endpoint
	if cfg.UseSSL && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + strings.TrimPrefix(endpoint, "http://")
	}
	if !cfg.UseSSL && !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to load rustfs aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = cfg.UsePathStyle
	})

	if _, err := client.ListBuckets(context.Background(), &s3.ListBucketsInput{}); err != nil {
		return fmt.Errorf("failed to connect rustfs: %w", err)
	}

	rustFS = &RustFS{
		client: client,
		bucket: cfg.Bucket,
		config: *cfg,
	}

	return nil
}

func GetRustFS() *RustFS {
	return rustFS
}

func (r *RustFS) Client() *s3.Client {
	if r == nil {
		return nil
	}
	return r.client
}

func (r *RustFS) Bucket() string {
	if r == nil {
		return ""
	}
	return r.bucket
}

func (r *RustFS) Config() config.RustFSConfig {
	if r == nil {
		return config.RustFSConfig{}
	}
	return r.config
}

func (r *RustFS) UploadObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) (string, string, error) {
	if r == nil || r.client == nil {
		return "", "", fmt.Errorf("rustfs is not initialized")
	}
	if r.bucket == "" {
		return "", "", fmt.Errorf("rustfs bucket is empty")
	}
	if key == "" {
		return "", "", fmt.Errorf("rustfs object key is empty")
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
		Body:   body,
	}
	if size >= 0 {
		input.ContentLength = aws.Int64(size)
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	if _, err := r.client.PutObject(ctx, input); err != nil {
		return "", "", fmt.Errorf("failed to upload object to rustfs: %w", err)
	}

	return key, r.objectURL(key), nil
}

func (r *RustFS) objectURL(key string) string {
	endpoint := strings.TrimSuffix(r.config.Endpoint, "/")
	if endpoint == "" {
		return ""
	}
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		if r.config.UseSSL {
			endpoint = "https://" + endpoint
		} else {
			endpoint = "http://" + endpoint
		}
	}

	base, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}

	if r.config.UsePathStyle {
		base.Path = path.Join(base.Path, r.bucket, key)
		return base.String()
	}

	base.Host = r.bucket + "." + base.Host
	base.Path = path.Join(base.Path, key)
	return base.String()
}

func (r *RustFS) DownloadObject(ctx context.Context, key string) ([]byte, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("rustfs is not initialized")
	}
	if r.bucket == "" {
		return nil, fmt.Errorf("rustfs bucket is empty")
	}
	if key == "" {
		return nil, fmt.Errorf("rustfs object key is empty")
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}

	output, err := r.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download object from rustfs: %w", err)
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	return data, nil
}
