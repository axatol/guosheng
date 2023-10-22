package discord

import "github.com/bwmarrin/discordgo"

type MessageEmbed discordgo.MessageEmbed

func NewMessageEmbed() *MessageEmbed {
	e := MessageEmbed{}
	e.Type = discordgo.EmbedTypeRich
	return &e
}

func (e *MessageEmbed) Embed() *discordgo.MessageEmbed {
	if e == nil {
		*e = MessageEmbed{}
	}

	if e.Type == "" {
		e.Type = discordgo.EmbedTypeRich
	}

	me := discordgo.MessageEmbed(*e)
	return &me
}

func (e *MessageEmbed) SetType(embedType discordgo.EmbedType) *MessageEmbed {
	e.Type = embedType
	return e
}

func (e *MessageEmbed) SetTitle(title string) *MessageEmbed {
	e.Title = title
	return e
}

func (e *MessageEmbed) SetURL(url string) *MessageEmbed {
	e.URL = url
	return e
}

func (e *MessageEmbed) AddField(name, value string, isInline ...bool) *MessageEmbed {
	inline := true
	if len(isInline) > 0 {
		inline = isInline[0]
	}

	e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})

	return e
}
