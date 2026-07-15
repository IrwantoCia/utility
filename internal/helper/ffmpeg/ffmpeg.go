// Package ffmpeg provides audio extraction from media files using FFmpeg.
// It converts any media file to WAV format suitable for speech-to-text
// processing: PCM 16-bit signed, 16kHz sample rate, mono channel.
package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ffmpeg CLI arguments
const (
	argInput      = "-i"
	argNoVideo    = "-vn"
	argAudioCodec = "-acodec"
	argSampleRate = "-ar"
	argChannels   = "-ac"
	argOverwrite  = "-y"
	argProgress   = "-progress"
	argNoStats    = "-nostats"
)

// ffmpeg audio settings
const (
	valPCM16LE      = "pcm_s16le"
	val16kHz        = "16000"
	valMono         = "1"
	valProgressPipe = "pipe:1"
)

// ffprobe CLI arguments
const (
	argLogLevel    = "-v"
	argShowEntries = "-show_entries"
	argOutputFmt   = "-of"
)

// ffprobe values
const (
	valError         = "error"
	valFormatDur     = "format=duration"
	valDefaultFormat = "default=noprint_wrappers=1:nokey=1"
)

// FFmpeg wraps the ffmpeg binary for audio extraction.
type FFmpeg struct {
	binPath string
}

// New finds ffmpeg in PATH and returns an FFmpeg instance.
func New() (*FFmpeg, error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	return &FFmpeg{binPath: path}, nil
}

// Convert extracts audio from inputPath to outputPath as WAV (PCM 16-bit, 16kHz, mono).
// If progress is non-nil, it receives real-time progress from 0.0 to 100.0.
// The context can be used to cancel the conversion.
func (f *FFmpeg) Convert(ctx context.Context, inputPath, outputPath string, progress func(percent float64)) error {
	if outputPath == "" {
		return ErrOutputPath
	}

	if progress == nil {
		cmd := exec.CommandContext(ctx,
			f.binPath,
			argInput, inputPath,
			argNoVideo,
			argAudioCodec, valPCM16LE,
			argSampleRate, val16kHz,
			argChannels, valMono,
			argOverwrite,
			outputPath,
		)

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("%w: %v", ErrConvert, err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("%w: %v", ErrConvert, err)
		}

		errOutput, _ := io.ReadAll(stderr)
		if err := cmd.Wait(); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("%w: %s", ErrConvert, string(errOutput))
		}

		return nil
	}

	// --- progress != nil ---

	totalDuration, err := getDuration(f.binPath, inputPath)
	if err != nil {
		totalDuration = 0
	}

	cmd := exec.CommandContext(ctx,
		f.binPath,
		argInput, inputPath,
		argNoVideo,
		argAudioCodec, valPCM16LE,
		argSampleRate, val16kHz,
		argChannels, valMono,
		argProgress, valProgressPipe,
		argNoStats,
		argOverwrite,
		outputPath,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConvert, err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConvert, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%w: %v", ErrConvert, err)
	}

	timeRegex := regexp.MustCompile(`^out_time_us=(\d+)$`)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if matches := timeRegex.FindStringSubmatch(line); matches != nil {
				us, _ := strconv.ParseInt(matches[1], 10, 64)
				if totalDuration > 0 {
					current := time.Duration(us) * time.Microsecond
					pct := float64(current) / float64(totalDuration) * 100.0
					if pct > 100.0 {
						pct = 100.0
					}
					progress(pct)
				}
			}
		}
	}()

	errOutput, _ := io.ReadAll(stderr)
	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%w: %s", ErrConvert, string(errOutput))
	}

	if totalDuration > 0 {
		progress(100.0)
	}

	return nil
}

// getDuration returns the duration of a media file using ffprobe.
func getDuration(ffmpegBin string, inputPath string) (time.Duration, error) {
	// ffprobe is typically alongside ffmpeg
	ffprobeBin := strings.Replace(ffmpegBin, "ffmpeg", "ffprobe", 1)

	cmd := exec.Command(
		ffprobeBin,
		argLogLevel, valError,
		argShowEntries, valFormatDur,
		argOutputFmt, valDefaultFormat,
		inputPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, err
	}

	return time.Duration(seconds * float64(time.Second)), nil
}
