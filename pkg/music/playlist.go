package music

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type Playlist interface {
	Advance() Song
	Enqueue(Song)
	RemoveByIndex(int)
	RemoveByID(string)
	Clear()
	Length() int
	List() []Song
}

var _ Playlist = (*playlistImpl)(nil)

func NewPlaylist() Playlist {
	return &playlistImpl{items: []Song{}}
}

type playlistImpl struct {
	mu    sync.RWMutex
	items []Song
}

func (pl *playlistImpl) Advance() Song {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	if len(pl.items) < 1 {
		return nil
	}

	song := pl.items[0]
	pl.items = pl.items[1:]
	return song
}

func (pl *playlistImpl) Enqueue(song Song) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	log.Trace().Str("song", song.ID()).Msg("enqueing song")
	pl.items = append(pl.items, song)
}

func (pl *playlistImpl) RemoveByIndex(i int) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	if i >= 0 && i < len(pl.items) {
		item := pl.items[i]
		log.Trace().Int("index", i).Str("song", item.ID()).Msg("removing item by index")
		pl.items = append(pl.items[0:i], pl.items[i+1:]...)
	}
}

func (pl *playlistImpl) RemoveByID(id string) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	for i, item := range pl.items {
		if item.ID() == id {
			log.Trace().Int("index", i).Str("song", item.ID()).Msg("removing item by id")
			pl.items = append(pl.items[0:i], pl.items[i+1:]...)
			return
		}
	}
}

func (pl *playlistImpl) Clear() {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.items = []Song{}
}

func (pl *playlistImpl) Length() int {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return len(pl.items)
}

func (pl *playlistImpl) List() []Song {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	items := make([]Song, len(pl.items))
	copy(items, pl.items)
	return items
}
