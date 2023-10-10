package discord

import (
	"context"
	"errors"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrCommandNotImplemented     = errors.New("command not implemented")
	ErrInvalidCommand            = errors.New("input did not satisfy a command interface")
	ErrInvalidApplicationCommand = errors.New("command does not implement the application command interface")
	ErrInvalidMessageCommand     = errors.New("command does not implement the message command interface")
)

type MessageCommandable interface {
	Name() string
	Description() string
	OnMessageCommand(context.Context, *Bot, *discordgo.MessageCreate, []string)
}

type ApplicationCommandable interface {
	ApplicationCommand() *discordgo.ApplicationCommand
	OnApplicationCommand(context.Context, *Bot, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionData)
}

func (b *Bot) RegisterCommand(ctx context.Context, cmd any) error {
	switch command := cmd.(type) {
	case MessageCommandable:
		b.Commands[command.Name()] = command

	case ApplicationCommandable:
		b.Commands[command.ApplicationCommand().Name] = command

	default:
		return ErrInvalidCommand
	}

	return nil
}
