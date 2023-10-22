package discord

import (
	"context"
	"errors"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrCommandNotImplemented                   = errors.New("command not implemented")
	ErrInvalidCommand                          = errors.New("input did not satisfy a command interface")
	ErrNotApplicationCommandInteractionHandler = errors.New("command does not implement the application command interaction handler interface")
	ErrNotMessageComponentInteractionHandler   = errors.New("command does not implement the message component interaction handler interface")
	ErrNotMessageHandler                       = errors.New("command does not implement the message handler interface")
)

type CommandMetadata interface {
	Name() string
	Description() string
}

type MessageHandler interface {
	CommandMetadata
	OnMessage(context.Context, *Bot, *discordgo.MessageCreate, []string)
}

type ApplicationCommandMetadata interface {
	ApplicationCommand() *discordgo.ApplicationCommand
}

type ApplicationCommandInteractionHandler interface {
	CommandMetadata
	ApplicationCommandMetadata
	OnApplicationCommand(context.Context, *Bot, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionData)
}

type MessageComponentInteractionHandler interface {
	CommandMetadata
	ApplicationCommandMetadata
	OnMessageComponent(context.Context, *Bot, *discordgo.InteractionCreate, *discordgo.MessageComponentInteractionData)
}

func (b *Bot) RegisterCommand(ctx context.Context, cmd any) error {
	if command, ok := cmd.(CommandMetadata); ok {
		b.Commands[command.Name()] = command
		return nil
	}

	return ErrInvalidCommand
}
