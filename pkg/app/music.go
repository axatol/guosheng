package app

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/axatol/guosheng/pkg/cache"
	"github.com/axatol/guosheng/pkg/music"
	"github.com/axatol/guosheng/pkg/yt"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func (a *App) GetPlayer(ctx context.Context, vc *discordgo.VoiceConnection) music.Player {
	id := fmt.Sprintf("%s:%s", vc.GuildID, vc.ChannelID)

	if _, ok := a.Players[id]; !ok {
		a.Players[id] = music.NewPlayer(ctx, id, vc.OpusSend, vc)
	}

	return a.Players[id]
}

func (a *App) GetSong(ctx context.Context, item *yt.Video) (song music.Song, err error) {
	cacheKey := fmt.Sprintf("cache/%s", item.ID)
	if _, err := a.ObjectStore.Stat(ctx, cacheKey); err != nil {
		if err != cache.ErrObjectNotFound {
			log.Error().Err(err).Send()
			return nil, err
		}

		raw, err := a.Executor.Download(item.ID)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, err
		}

		if _, err := a.ObjectStore.Put(ctx, cacheKey, raw, item.ToMap()); err != nil {
			log.Error().Err(err).Send()
			return nil, err
		}
	}

	raw, err := a.ObjectStore.Get(ctx, cacheKey)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, err
	}

	raw, err = a.Executor.Encode(item.ID, raw)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, err
	}

	buffer := bytes.NewBuffer(raw)
	frames := [][]byte{}

	for {
		var frameLength int16
		if err := binary.Read(buffer, binary.LittleEndian, &frameLength); err != nil {
			return nil, fmt.Errorf("failed to read frame length: %s", err)
		}

		frame := make([]byte, frameLength)
		if err := binary.Read(buffer, binary.LittleEndian, &frame); err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				return nil, fmt.Errorf("failed to read frame: %s", err)
			}

			break
		}

		frames = append(frames, frame)
	}

	song = music.NewSong(item.ID, frames)
	return song, nil
}

func (a *App) PlaySong(ctx context.Context, vc *discordgo.VoiceConnection, song music.Song) {
	p := a.GetPlayer(ctx, vc)
	p.Playlist().Enqueue(song)
	p.Play()
}
