package cmds

import "github.com/bwmarrin/discordgo"

func resolveOptions(opts []*discordgo.ApplicationCommandInteractionDataOption) map[string]any {
	result := map[string]any{}

	for _, opt := range opts {
		switch opt.Type {
		case discordgo.ApplicationCommandOptionString:
			result[opt.Name] = opt.StringValue()
		case discordgo.ApplicationCommandOptionBoolean:
			result[opt.Name] = opt.BoolValue()
		case discordgo.ApplicationCommandOptionInteger:
			result[opt.Name] = opt.IntValue()
		case discordgo.ApplicationCommandOptionSubCommand, discordgo.ApplicationCommandOptionSubCommandGroup:
			result[opt.Name] = resolveOptions(opt.Options)
		case discordgo.ApplicationCommandOptionNumber:
			result[opt.Name] = opt.FloatValue()
		}
	}

	return result
}
