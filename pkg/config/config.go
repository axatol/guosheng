package config

import (
	"flag"
	"os"
	"runtime"
	"strings"

	"github.com/axatol/go-utils/flags"
	"github.com/axatol/go-utils/ptr"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	buildCommit  string = "unknown"
	buildTime    string = "unknown"
	PrintVersion bool

	logLevel       = flags.LogLevelValue{Default: ptr.Ptr(zerolog.InfoLevel.String())}
	logFormat      = flags.EnumValue{Valid: []string{"json", "text"}, Default: ptr.Ptr("json")}
	configFilename string

	DiscordAppID         string
	DiscordBotToken      string
	DiscordMessagePrefix string

	ServerAddress string

	YouTubeAPIKey string
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
	fs.StringVar(&ServerAddress, "server-address", ":8080", "server address")
	fs.StringVar(&YouTubeAPIKey, "youtube-api-key", ":8080", "youtube api key")

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
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
}
