package config

import (
	"flag"
	"os"
	"runtime"
	"strings"

	"github.com/axatol/go-utils/flags"
	"github.com/axatol/go-utils/ptr"
	"github.com/axatol/guosheng/pkg/util"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	buildCommit    string = "unknown"
	buildTime      string = "unknown"
	PrintVersion   bool
	logLevel       = flags.LogLevelValue{Default: ptr.Ptr(zerolog.InfoLevel.String())}
	logFormat      = flags.EnumValue{Valid: []string{"json", "text"}, Default: ptr.Ptr("json")}
	configFilename string

	DiscordAppID         string
	DiscordBotToken      string
	DiscordMessagePrefix string

	MinioEnabled         bool
	MinioEndpoint        string
	MinioBucket          string
	MinioAccessKeyID     string
	MinioSecretAccessKey string

	ServerAddress string

	YouTubeAPIKey string

	YTDLPExecutable     string
	DCAExecutable       string
	FFMPEGExecutable    string
	YTDLPCacheDirectory string
	YTDLPConcurrency    int
)

func Version() *zerolog.Logger {
	e := log.With().
		Str("go_os", runtime.GOOS).
		Str("go_arch", runtime.GOARCH).
		Str("go_version", runtime.Version()).
		Str("build_commit", buildCommit).
		Str("build_time", buildTime).
		Str("discordgo_version", discordgo.VERSION).
		Logger()
	return &e
}

func Configure() {
	fs := flags.FlagSet{FlagSet: flag.CommandLine}
	fs.Var(&logLevel, "log-level", "log level")
	fs.Var(&logFormat, "log-format", "log format")
	fs.BoolVar(&PrintVersion, "version", false, "prints the program version")
	fs.StringVar(&configFilename, "config", "", "config file name")

	fs.StringVar(&DiscordAppID, "discord-app-id", "", "discord app id")
	fs.StringVar(&DiscordBotToken, "discord-bot-token", "", "discord bot token")
	fs.StringVar(&DiscordMessagePrefix, "discord-message-prefix", "", "discord message prefix")

	fs.BoolVar(&MinioEnabled, "minio-enabled", false, "minio enabled")
	fs.StringVar(&MinioEndpoint, "minio-endpoint", "", "minio endpoint")
	fs.StringVar(&MinioBucket, "minio-bucket", "", "minio bucket")
	fs.StringVar(&MinioAccessKeyID, "minio-access-key-id", "", "minio access key id")
	fs.StringVar(&MinioSecretAccessKey, "minio-secret-access-key", "", "minio secret access key")

	fs.StringVar(&ServerAddress, "server-address", ":8080", "server address")

	fs.StringVar(&YouTubeAPIKey, "youtube-api-key", "", "youtube api key")

	fs.StringVar(&YTDLPExecutable, "ytdlp-executable", "yt-dlp", "yt-dlp executable")
	fs.StringVar(&DCAExecutable, "dca-executable", "dca", "dca executable")
	fs.StringVar(&FFMPEGExecutable, "ffmpeg-executable", "ffmpeg", "ffmpeg executable")
	fs.StringVar(&YTDLPCacheDirectory, "ytdlp-cache-directory", "/var/cache/ytdlp", "yt-dlp cache directory")
	fs.IntVar(&YTDLPConcurrency, "ytdlp-concurrency", 3, "yt dlp concurrency")

	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	godotenv.Load()
	if err := fs.LoadUnsetFromEnv(); err != nil {
		panic(err)
	}

	if configFilename != "" && strings.HasSuffix(configFilename, "json") {
		if err := fs.LoadUnsetFromJSONFile(configFilename); err != nil {
			panic(err)
		}
	}

	if configFilename != "" && (strings.HasSuffix(configFilename, "yaml") || strings.HasSuffix(configFilename, "yml")) {
		if err := fs.LoadUnsetFromYAMLFile(configFilename); err != nil {
			panic(err)
		}
	}

	zerologLevel, err := zerolog.ParseLevel(logLevel.String())
	if err != nil {
		panic(err)
	}

	zerolog.SetGlobalLevel(zerologLevel)
	if logFormat.String() == "text" {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	log.Logger = log.Logger.With().
		Timestamp().
		Stack().
		Caller().
		Logger()

	Version().Debug().
		Int("pid", os.Getpid()).
		Str("log_level", logLevel.String()).
		Str("log_format", logFormat.String()).
		Str("config_filename", configFilename).
		Str("discord_app_id", DiscordAppID).
		Str("discord_bot_token", util.Obscure(DiscordBotToken, 3)).
		Str("discord_message_prefix", DiscordMessagePrefix).
		Bool("minio_enabled", MinioEnabled).
		Str("minio_endpoint", MinioEndpoint).
		Str("minio_bucket", MinioBucket).
		Str("minio_access_key_id", util.Obscure(MinioAccessKeyID, 3)).
		Str("minio_secret_access_key", util.Obscure(MinioSecretAccessKey, 3)).
		Str("server_address", ServerAddress).
		Str("youtube_api_key", util.Obscure(YouTubeAPIKey, 3)).
		Str("ytdlp_executable", YTDLPExecutable).
		Str("dca_executable", DCAExecutable).
		Str("ffmpeg_executable", FFMPEGExecutable).
		Str("ytdlp_cache_directory", YTDLPCacheDirectory).
		Int("ytdlp_concurrency", YTDLPConcurrency).
		Send()
}
