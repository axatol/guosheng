package cache

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrObjectNotFound = errors.New("object not found")
)

type ObjectInfo struct {
	Key  string
	ETag string
	Size int64
	Tags map[string]string
}

type ObjectStoreType string

const (
	MinioObjectStore      ObjectStoreType = "minio"
	FilesystemObjectStore ObjectStoreType = "filesystem"
)

func (t ObjectStoreType) String() string {
	return string(t)
}

type ObjectStore interface {
	Type() ObjectStoreType
	Get(context.Context, string) ([]byte, error)
	Put(context.Context, string, []byte, map[string]string) (*ObjectInfo, error)
	Stat(context.Context, string) (*ObjectInfo, error)
}

type ObjectStoreOptions struct {
	filesystem *FilesystemClientOptions
	minio      *MinioClientOptions
}

func (o *ObjectStoreOptions) SetFilesystem(baseDir string) *ObjectStoreOptions {
	o.filesystem = &FilesystemClientOptions{
		BaseDir: baseDir,
	}

	return o
}

func (o *ObjectStoreOptions) SetMinio(endpoint, bucket, accessKeyID, secretAccessKey string) *ObjectStoreOptions {
	o.minio = &MinioClientOptions{
		Endpoint:        endpoint,
		Bucket:          bucket,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}

	return o
}

func NewObjectStore(opts ObjectStoreOptions) (ObjectStore, error) {
	if opts.minio != nil {
		return NewMinioClient(*opts.minio)
	}

	if opts.filesystem != nil {
		return NewFilesystemClient(*opts.filesystem), nil
	}

	return nil, fmt.Errorf("no options configured")
}
