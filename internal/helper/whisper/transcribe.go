package whisper

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

// Whisper wraps the whisper-cli binary for speech-to-text.
type Whisper struct {
	binPath string
}

// New finds whisper-cli in PATH and returns a Whisper instance.
func New() (*Whisper, error) {
	path, err := exec.LookPath("whisper-cli")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTranscribe, err)
	}
	return &Whisper{binPath: path}, nil
}

// Transcribe runs speech-to-text on the input audio file.
//
// Parameters:
//   - ctx: cancellable context
//   - modelPath: full path to the ggml-*.bin model file
//   - inputPath: path to the WAV/PCM audio file (must be 16kHz, mono, 16-bit)
//   - outputBase: base output filename WITHOUT extension (e.g., "filmku" produces "filmku.txt")
//   - language: language code ("auto", "en", "id", "ja", "tl") or "auto"
//   - formats: output formats ("txt", "srt", "vtt", "csv", "json", "lrc"). Empty defaults to ["txt"].
//   - progress: called every ~1 second with elapsed time. nil = no progress reporting.
//     Caller MUST NOT block in this callback (send to channel instead).
func (w *Whisper) Transcribe(ctx context.Context, modelPath, inputPath, outputBase, language, vadModelPath string, formats []string, progress func(elapsed time.Duration)) error {
	args := []string{
		"-m", modelPath,
		"-f", inputPath,
		"-l", language,
		"-of", outputBase,
	}

	if len(formats) == 0 {
		formats = []string{"txt"}
	}
	for _, f := range formats {
		args = append(args, "-o"+f)
	}

	if vadModelPath != "" {
		args = append(args, "--vad", "--vad-model", vadModelPath)
	}

	cmd := exec.CommandContext(ctx, w.binPath, args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTranscribe, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%w: %v", ErrTranscribe, err)
	}

	if progress != nil {
		start := time.Now()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		done := make(chan struct{})
		defer close(done)

		go func() {
			for {
				select {
				case <-ticker.C:
					progress(time.Since(start))
				case <-done:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	errOutput, _ := io.ReadAll(stderr)
	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%w: %s", ErrTranscribe, string(errOutput))
	}

	return nil
}
