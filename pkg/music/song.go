package music

import "fmt"

type Song interface {
	ID() string
	Frames() [][]byte
	Frame() []byte
	HasFrame() bool
	Progress() int
	String() string
}

var _ Song = (*songImpl)(nil)

func NewSong(id string, frames [][]byte) Song {
	return &songImpl{id, frames, 0}
}

type songImpl struct {
	id       string
	frames   [][]byte
	progress int
}

func (s *songImpl) ID() string {
	return s.id
}

func (s *songImpl) Frames() [][]byte {
	return s.frames
}

func (s *songImpl) Frame() []byte {
	frame := s.frames[s.progress]
	s.progress += 1
	return frame
}

func (s *songImpl) HasFrame() bool {
	return s.progress < len(s.frames)
}

func (s *songImpl) Progress() int {
	return s.progress
}

func (s *songImpl) String() string {
	return fmt.Sprintf("%s(%d/%d)", s.id, s.progress+1, len(s.frames))
}
