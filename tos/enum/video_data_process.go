package enum

// VideoDataProcessStatus represents the status field returned by video data processing APIs.
// Currently only "OK" is observed.
type VideoDataProcessStatus string

const (
	VideoDataProcessStatusOK VideoDataProcessStatus = "OK"
)

