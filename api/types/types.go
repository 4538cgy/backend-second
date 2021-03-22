package types

type MediaType int

const (
	ImageType = MediaType(0)
	VideoType = MediaType(1)
)

func (m MediaType) String() string {
	switch m {
	case ImageType:
		return "Image"
	case VideoType:
		return "Video"
	}
	return "Unknown"
}

type (
	MediaInfo struct {
		Kind       string `json:"media_kind"`
		MediaId    string `json:"media_id"`
		MediaUrl   string `json:"media_url"`
		ServeReady int    `json:"serve_ready"`
	}
	MediaInfos struct {
		Item []MediaInfo `json:"media_infos"`
	}
)

type MediaIndices struct {
	MediaIds []string `json:"media_indices"`
}
