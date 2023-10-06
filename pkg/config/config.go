package config

import (
	"flag"
	"os"
	"strings"

	"github.com/axatol/go-utils/flags"
	"github.com/axatol/go-utils/ptr"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	logLevel       = flags.LogLevelValue{Default: ptr.Ptr(zerolog.InfoLevel.String())}
	logFormat      = flags.EnumValue{Valid: []string{"json", "text"}, Default: ptr.Ptr("json")}
	configFilename string

	DiscordBotPrefix string
	DiscordBotToken  string

	ServerAddress string
)

func Configure() {
	fs := flags.FlagSet{FlagSet: flag.CommandLine}
	fs.Var(&logLevel, "log-level", "log level")
	fs.Var(&logFormat, "log-format", "log format")
	fs.StringVar(&configFilename, "config", "", "config file name")
	fs.StringVar(&DiscordBotPrefix, "discord-bot-prefix", "", "discord bot prefix")
	fs.StringVar(&DiscordBotToken, "discord-bot-token", "", "discord bot token")
	fs.StringVar(&ServerAddress, "server-address", ":8080", "server address")

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
