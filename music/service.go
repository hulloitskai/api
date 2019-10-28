package music // import "go.stevenxie.me/api/v2/music"

type (
	// A Service handles all music-related requests.
	Service interface {
		SourceService
		CurrentService
		ControlService
	}

	// A Streamer handles all music-related streams.
	Streamer interface {
		CurrentStreamer
	}
)
