package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type CommandMetadata interface {
	Name() string
	Description() string
}

type Commandable interface {
	CommandMetadata
	OnMessage(context.Context, *Bot, *discordgo.MessageCreate, []string) error
}

type Interactable interface {
	CommandMetadata
	Interaction() *discordgo.ApplicationCommand
}

type ApplicationCommandInteractive interface {
	Interactable
	OnApplicationCommandInteraction(context.Context, *Bot, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionData) error
}

// type MessageComponentInteractive interface {
// 	Interactable
// 	OnMessageComponentInteraction(context.Context, *Bot, *discordgo.InteractionCreate, *discordgo.MessageComponentInteractionData) error
// }

// type ModalSubmitInteractive interface {
// 	Interactable
// 	OnModalSubmitInteraction(context.Context, *Bot, *discordgo.InteractionCreate, *discordgo.ModalSubmitInteractionData) error
// }

func (b *Bot) RegisterCommand(ctx context.Context, cmd any) error {
	metadata, ok := cmd.(CommandMetadata)
	if !ok {
		return fmt.Errorf("input did not satisfy CommandMetadata")
	}

	if command, ok := cmd.(Commandable); ok {
		b.commands[metadata.Name()] = command
	}

	if interactive, ok := cmd.(Interactable); ok {
		if _, err := b.Session.ApplicationCommandCreate(b.AppID, "", interactive.Interaction(), discordgo.WithContext(ctx)); err != nil {
			return fmt.Errorf("failed to create application command %s: %s", interactive.Name(), err)
		}

		b.interactions[metadata.Name()] = interactive
	}

	return nil
}
