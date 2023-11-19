package music_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/axatol/guosheng/pkg/music"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.RFC3339
		w.Out = os.Stdout
		w.NoColor = true
	})).With().Timestamp().Caller().Logger()
}

type mockSpeaker struct {
	startedSpeaking bool
	stoppedSpeaking bool
}

func (s *mockSpeaker) Speaking(state bool) error {
	if state {
		s.startedSpeaking = true
	} else {
		s.stoppedSpeaking = true
	}

	return nil
}

var (
	eDeadline = time.Second
	eTick     = time.Millisecond * 10
)

func TestPlayerPlay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	t.Cleanup(cancel)

	song := newSong("songid")
	dest := make(chan []byte)
	speaker := &mockSpeaker{}
	player := music.NewPlayer(ctx, "playerid", dest, speaker)

	// enqueue songs
	player.Playlist().Enqueue(song)
	assert.Len(t, player.Playlist().List(), 1)

	// start playing
	player.Play()
	assert.EventuallyWithT(t, func(t *assert.CollectT) { assert.NotNil(t, player.Current()) }, eDeadline, eTick)
	assert.True(t, speaker.startedSpeaking)

	// start consuming, ensuring order
	for _, frame := range song.Frames() {
		assert.ElementsMatch(t, frame, <-dest)
	}
	assert.EventuallyWithT(t, func(t *assert.CollectT) { assert.True(t, speaker.stoppedSpeaking) }, eDeadline, eTick)
}

func TestPlayerPause(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	t.Cleanup(cancel)

	dest := make(chan []byte)
	speaker := &mockSpeaker{}
	player := music.NewPlayer(ctx, "playerid", dest, speaker)

	// enqueue songs
	for i := 0; i < 3+rand.Intn(7); i++ {
		player.Playlist().Enqueue(newSong(fmt.Sprintf("songid%d", i)))
	}

	consume := func(c chan []byte) bool {
		select {
		case _, ok := <-c:
			if !ok {
				log.Debug().Msg("consume: channel closed")
				return false
			}
			log.Debug().Msg("consume: got something")
			return true
		default:
			log.Debug().Msg("consume: nothing coming")
			return false
		}
	}

	// begin playback
	player.Play()
	assert.EventuallyWithT(t, func(t *assert.CollectT) { assert.True(t, consume(dest)) }, eDeadline, eTick)

	// pause playback
	player.Pause()
	assert.EventuallyWithT(t, func(t *assert.CollectT) { assert.False(t, consume(dest)) }, eDeadline, eTick)

	// resume playback
	player.Play()
	assert.EventuallyWithT(t, func(t *assert.CollectT) { assert.True(t, consume(dest)) }, eDeadline, eTick)
}
