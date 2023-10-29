package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
)

var (
	_ ObjectStore = (*FilesystemClient)(nil)
)

type FilesystemClientOptions struct {
	BaseDir string
}

func NewFilesystemClient(opts FilesystemClientOptions) *FilesystemClient {
	return &FilesystemClient{baseDir: opts.BaseDir}
}

type FilesystemClient struct {
	baseDir string
}

func (c *FilesystemClient) Type() ObjectStoreType {
	return FilesystemObjectStore
}

func (c *FilesystemClient) Get(ctx context.Context, key string) ([]byte, error) {
	filename := path.Join(c.baseDir, key)
	file, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrObjectNotFound
		}

		return nil, fmt.Errorf("failed to read file %s: %s", filename, err)
	}

	return file, nil
}

func (c *FilesystemClient) Put(ctx context.Context, key string, raw []byte, tags map[string]string) (*ObjectInfo, error) {
	filename := path.Join(c.baseDir, key)
	if err := os.WriteFile(filename, raw, 0666); err != nil {
		return nil, fmt.Errorf("failed to write file %s: %s", filename, err)
	}

	if tags != nil {
		tagsFilename := fmt.Sprintf("%s.tags.json", filename)
		tagsRaw, err := json.Marshal(tags)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal file tags %s: %s", tagsFilename, err)
		}

		if err := os.WriteFile(tagsFilename, tagsRaw, 0666); err != nil {
			return nil, fmt.Errorf("failed to write tag file %s: %s", tagsFilename, tagsRaw)
		}
	}

	info := ObjectInfo{
		Key:  key,
		ETag: key,
		Size: int64(len(raw)),
		Tags: tags,
	}

	return &info, nil
}

func (c *FilesystemClient) Stat(ctx context.Context, key string) (*ObjectInfo, error) {
	filename := path.Join(c.baseDir, key)
	stat, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrObjectNotFound
		}

		return nil, fmt.Errorf("failed to stat file %s: %s", filename, err)
	}

	tagsFilename := fmt.Sprintf("%s.tags.json", filename)
	raw, err := os.ReadFile(tagsFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to read tag file %s: %s", tagsFilename, err)
	}

	var tags map[string]string
	if raw != nil {
		if err := json.Unmarshal(raw, &tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tag file %s: %s", tagsFilename, err)
		}
	}

	info := ObjectInfo{
		Key:  key,
		ETag: key,
		Size: stat.Size(),
		Tags: tags,
	}

	return &info, nil
}
