package music_test

import (
	"testing"

	"github.com/axatol/guosheng/pkg/music"
	"github.com/stretchr/testify/assert"
)

func TestPlaylistEnqueueAdvance(t *testing.T) {
	playlist := music.NewPlaylist()
	song0ID := "song0"
	song1ID := "song1"
	song2ID := "song2"
	playlist.Enqueue(newSong(song0ID))
	playlist.Enqueue(newSong(song1ID))
	playlist.Enqueue(newSong(song2ID))
	assert.Len(t, playlist.List(), 3)
	assert.Equal(t, playlist.List()[0].ID(), song0ID)
	assert.Equal(t, playlist.List()[1].ID(), song1ID)
	assert.Equal(t, playlist.List()[2].ID(), song2ID)
	song0 := playlist.Advance()
	assert.NotNil(t, song0)
	assert.Equal(t, song0.ID(), song0ID)
	song1 := playlist.Advance()
	assert.NotNil(t, song1)
	assert.Equal(t, song1.ID(), song1ID)
	song2 := playlist.Advance()
	assert.NotNil(t, song2)
	assert.Equal(t, song2.ID(), song2ID)
	assert.Nil(t, playlist.Advance())
	assert.Empty(t, playlist.List())
}

func TestPlaylistRemoveByIndex(t *testing.T) {
	playlist := music.NewPlaylist()
	song0ID := "song0"
	song1ID := "song1"
	song2ID := "song2"
	playlist.Enqueue(newSong(song0ID))
	playlist.Enqueue(newSong(song1ID))
	playlist.Enqueue(newSong(song2ID))
	assert.Len(t, playlist.List(), 3)
	playlist.RemoveByIndex(1)
	assert.Len(t, playlist.List(), 2)
	assert.Equal(t, playlist.List()[0].ID(), song0ID)
	assert.Equal(t, playlist.List()[1].ID(), song2ID)
}

func TestPlaylistRemoveByID(t *testing.T) {
	playlist := music.NewPlaylist()
	song0ID := "song0"
	song1ID := "song1"
	song2ID := "song2"
	playlist.Enqueue(newSong(song0ID))
	playlist.Enqueue(newSong(song1ID))
	playlist.Enqueue(newSong(song2ID))
	assert.Len(t, playlist.List(), 3)
	playlist.RemoveByID(song1ID)
	assert.Len(t, playlist.List(), 2)
	assert.Equal(t, playlist.List()[0].ID(), song0ID)
	assert.Equal(t, playlist.List()[1].ID(), song2ID)
}
