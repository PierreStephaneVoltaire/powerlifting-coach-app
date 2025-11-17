package processing

type VideoMetadata struct {
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Resolution string  `json:"resolution"`
	Codec      string  `json:"codec"`
	Bitrate    int     `json:"bitrate"`
	Framerate  float64 `json:"framerate"`
}
