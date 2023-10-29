package cache

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	_ ObjectStore = (*MinioClient)(nil)
)

type MinioClientOptions struct {
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
}

func NewMinioClient(opts MinioClientOptions) (*MinioClient, error) {
	secure := strings.HasPrefix(opts.Endpoint, "https")

	endpoint := opts.Endpoint
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(opts.AccessKeyID, opts.SecretAccessKey, ""),
		Secure: secure,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create minio client for endpoint %s: %s", endpoint, err)
	}

	return &MinioClient{client, opts.Bucket}, err
}

type MinioClient struct {
	client     *minio.Client
	bucketName string
}

func (c *MinioClient) Type() ObjectStoreType {
	return MinioObjectStore
}

func (c *MinioClient) Get(ctx context.Context, key string) ([]byte, error) {
	object, err := c.client.GetObject(ctx, c.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, ErrObjectNotFound
		}

		return nil, fmt.Errorf("failed to get object %s: %s", key, err)
	}

	raw, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object %s: %s", key, err)
	}

	return raw, nil
}

func (c *MinioClient) Put(ctx context.Context, key string, raw []byte, tags map[string]string) (*ObjectInfo, error) {
	reader := bytes.NewReader(raw)
	upload, err := c.client.PutObject(ctx, c.bucketName, key, reader, int64(reader.Len()), minio.PutObjectOptions{UserTags: tags})
	if err != nil {
		return nil, fmt.Errorf("failed to put object %s: %s", key, err)
	}

	info := ObjectInfo{
		Key:  upload.Key,
		ETag: upload.ETag,
		Size: upload.Size,
		Tags: tags,
	}

	return &info, nil
}

func (c *MinioClient) Stat(ctx context.Context, key string) (*ObjectInfo, error) {
	stat, err := c.client.StatObject(ctx, c.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, ErrObjectNotFound
		}

		return nil, fmt.Errorf("failed to stat object %s: %s", key, err)
	}

	info := ObjectInfo{
		Key:  stat.Key,
		ETag: stat.ETag,
		Size: stat.Size,
		Tags: stat.UserTags,
	}

	return &info, nil
}
