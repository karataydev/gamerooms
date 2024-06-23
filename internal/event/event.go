package event

type Event string

const (
	Connect Event = "Connect"
	Ready Event = "Ready"
	Vote Event = "Vote"
)


const (
	InvalidEvent Event = "InvalidEvent"
	UnsupportedEvent Event = "UnsupportedEvent"
)
