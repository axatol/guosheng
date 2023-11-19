package music

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type speakable interface {
	Speaking(bool) error
}

type Player interface {
	ID() string
	Playing() bool
	Current() Song
	Playlist() Playlist
	Play()
	Pause()
	Next()
}

var _ Player = (*playerImpl)(nil)

func NewPlayer(ctx context.Context, id string, dest chan []byte, speakable speakable) Player {
	player := playerImpl{}
	player.id = id
	player.dest = dest
	player.frames = make(chan []byte)
	player.controls = make(chan control)
	player.speakable = speakable
	player.playlist = NewPlaylist()

	pauser := sync.Mutex{}

	go func() {
		for frame := range player.frames {
			pauser.Lock()
			log.Trace().Msg("sending song frame")
			player.dest <- frame
			pauser.Unlock()
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Trace().Msg("player stopping")
				close(player.frames)
				return

			case ctl := <-player.controls:
				// CTL:
				log.Trace().Str("control", ctl.String()).Send()
				switch ctl {
				case nextControl:
					pauser.TryLock()
					pauser.Unlock()
					player.startPlaying()
					player.nextSong()

				case playControl:
					pauser.TryLock()
					pauser.Unlock()
					player.startPlaying()
					if player.current == nil {
						player.nextSong()
					}

				case pauseControl:
					pauser.Lock()
					player.stopPlaying()
				}

			default:
				if !player.Playing() {
					log.Trace().Msg("sleeping: paused")
					time.Sleep(time.Millisecond * 500)
					continue
				}

				if player.Current() == nil && player.Playlist().Length() < 1 {
					log.Trace().Msg("sleeping: playlist empty")
					time.Sleep(time.Millisecond * 500)
					continue
				}

				if player.Current() == nil {
					log.Trace().Msg("advancing playlist: no current song")
					player.nextSong()
					continue
				}

				if !player.Current().HasFrame() {
					log.Trace().Msg("advancing playlist: song finished")
					player.stopPlaying()
					player.nextSong()
					continue
				}

				log.Trace().Str("song", player.Current().String()).Msg("buffering song frame")
				player.frames <- player.Current().Frame()
			}
		}
	}()

	return &player
}

type playerImpl struct {
	mu        sync.RWMutex // use to lock playing, frames, current
	id        string
	playing   bool
	current   Song
	playlist  Playlist
	controls  chan control
	frames    chan []byte
	dest      chan []byte
	speakable speakable
}

func (p *playerImpl) ID() string {
	return p.id
}

func (p *playerImpl) Playing() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.playing
}

func (p *playerImpl) Current() Song {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.current
}

func (p *playerImpl) Playlist() Playlist {
	return p.playlist
}

func (p *playerImpl) Play() {
	p.controls <- playControl
}

func (p *playerImpl) Pause() {
	p.controls <- pauseControl
}

func (p *playerImpl) Next() {
	p.controls <- nextControl
}

func (p *playerImpl) startPlaying() {
	log.Trace().Msg("start playing")

	p.mu.Lock()
	p.playing = true
	p.mu.Unlock()

	if err := p.speakable.Speaking(true); err != nil {
		log.Error().Err(fmt.Errorf("failed to start speaking: %s", err)).Send()
	}
}

func (p *playerImpl) stopPlaying() {
	log.Trace().Msg("stop playing")

	p.mu.Lock()
	p.playing = false
	p.mu.Unlock()

	if err := p.speakable.Speaking(false); err != nil {
		log.Error().Err(fmt.Errorf("failed to stop speaking: %s", err)).Send()
	}
}

func (p *playerImpl) nextSong() {
	log.Trace().Msg("next song")

	p.mu.Lock()
	p.current = p.playlist.Advance()
	p.mu.Unlock()
}
