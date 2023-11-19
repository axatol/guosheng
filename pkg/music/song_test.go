package music_test

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"testing"

	"github.com/axatol/guosheng/pkg/music"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func newSong(id string) music.Song {
	frames := make([][]byte, rand.Intn(7)+3)
	for i := range frames {
		frames[i] = make([]byte, rand.Intn(7)+3)
		if _, err := crand.Read(frames[i]); err != nil {
			log.Fatal().Err(fmt.Errorf("failed to create song frame: %s", err)).Send()
		}
	}

	return music.NewSong(id, frames)
}

func TestSong(t *testing.T) {
	song := newSong("song")

	for i, frame := range song.Frames() {
		assert.Equal(t, i, song.Progress())
		assert.True(t, song.HasFrame())
		assert.Equal(t, frame, song.Frame())
	}

	assert.False(t, song.HasFrame())
	assert.Equal(t, len(song.Frames()), song.Progress())
}
