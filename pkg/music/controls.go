package music

type control int

const (
	playControl control = iota
	pauseControl
	nextControl
)

var controls = [...]string{
	playControl:  "play",
	pauseControl: "pause",
	nextControl:  "next",
}

func (t control) String() string {
	if t < 0 || int(t) >= len(controls) {
		return "unknown"
	}

	return controls[t]
}
