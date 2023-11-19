package app

import (
	"context"

	"github.com/axatol/guosheng/pkg/cache"
	"github.com/axatol/guosheng/pkg/cli"
	"github.com/axatol/guosheng/pkg/config"
	"github.com/axatol/guosheng/pkg/music"
	"github.com/axatol/guosheng/pkg/yt"
)

type App struct {
	YouTube     *yt.Client
	Executor    *cli.Executor
	Players     map[string]music.Player
	ObjectStore cache.ObjectStore
}

func New(ctx context.Context) (*App, error) {
	youtube, err := yt.New(ctx, config.YouTubeAPIKey)
	if err != nil {
		return nil, err
	}

	executor := cli.Executor{
		YTDLPExecutable:  config.YTDLPExecutable,
		FFMPEGExecutable: config.FFMPEGExecutable,
		DCAExecutable:    config.DCAExecutable,
		Concurrency:      config.YTDLPConcurrency,
		CacheDirectory:   config.YTDLPCacheDirectory,
	}

	if err := executor.Listen(ctx); err != nil {
		return nil, err
	}

	objectStoreOpts := cache.ObjectStoreOptions{}
	objectStoreOpts.SetFilesystem(config.YTDLPCacheDirectory)
	if config.MinioEnabled {
		objectStoreOpts.SetMinio(config.MinioEndpoint, config.MinioBucket, config.MinioAccessKeyID, config.MinioSecretAccessKey)
	}

	objectStore, err := cache.NewObjectStore(objectStoreOpts)
	if err != nil {
		return nil, err
	}

	app := App{
		YouTube:     youtube,
		Executor:    &executor,
		ObjectStore: objectStore,
		Players:     map[string]music.Player{},
	}

	return &app, nil
}
