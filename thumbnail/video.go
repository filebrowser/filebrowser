package thumbnail

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

var ErrUnavailable = errors.New("thumbnail generator unavailable")

type Video struct {
	ffmpeg  string
	ffprobe string
}

func NewVideo() *Video {
	ffmpeg, _ := exec.LookPath("ffmpeg")
	ffprobe, _ := exec.LookPath("ffprobe")
	return NewVideoWithTools(ffmpeg, ffprobe)
}

func NewVideoWithTools(ffmpeg, ffprobe string) *Video {
	return &Video{
		ffmpeg:  ffmpeg,
		ffprobe: ffprobe,
	}
}

func (v *Video) Thumbnail(ctx context.Context, input string, out io.Writer) error {
	if v == nil || v.ffmpeg == "" || v.ffprobe == "" {
		return ErrUnavailable
	}

	duration, err := v.duration(ctx, input)
	if err != nil {
		if ctx.Err() != nil {
			return err
		}
		duration = 0
	}

	args := []string{
		"-nostdin",
		"-hide_banner",
		"-loglevel", "error",
		"-ss", seekOffset(duration),
		"-i", input,
		"-frames:v", "1",
		"-vf", "scale=256:256:force_original_aspect_ratio=increase,crop=256:256",
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"pipe:1",
	}

	cmd := exec.CommandContext(ctx, v.ffmpeg, args...)
	cmd.Stdout = out
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	_, readErr := io.ReadAll(stderr)
	waitErr := cmd.Wait()
	if waitErr != nil {
		return fmt.Errorf("ffmpeg thumbnail failed: %w", waitErr)
	}

	return readErr
}

func (v *Video) duration(ctx context.Context, input string) (float64, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	}

	out, err := exec.CommandContext(ctx, v.ffprobe, args...).Output()
	if err != nil {
		return 0, err
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

func seekOffset(duration float64) string {
	if duration <= 0 {
		return "0.100"
	}
	if duration < 1 {
		return fmt.Sprintf("%.3f", duration/2)
	}

	seek := duration * 0.1
	if seek < 0.1 {
		seek = 0.1
	}
	if seek > 10 {
		seek = 10
	}
	if seek >= duration {
		seek = duration / 2
	}

	return fmt.Sprintf("%.3f", seek)
}
