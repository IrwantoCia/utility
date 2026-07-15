package whisper

import "errors"

// Sentinel errors for the whisper package.
var (
	ErrTranscribe = errors.New("whisper: transcription failed")
)
