//go:generate go-enum --sql --marshal --file $GOFILE
package img

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/marusama/semaphore/v2"
	"github.com/spf13/afero"
)

// ErrUnsupportedFormat means the given image format is not supported.
var ErrUnsupportedFormat = errors.New("unsupported image format")

// Service
type Service struct {
	lowPrioritySem  semaphore.Semaphore
	highPrioritySem semaphore.Semaphore
}

func New(workers int) *Service {
	return &Service{
		lowPrioritySem:  semaphore.New(workers),
		highPrioritySem: semaphore.New(workers),
	}
}

// Format is an image file format.
/*
ENUM(
jpeg
png
gif
tiff
bmp
)
*/
type Format int

func (x Format) toImaging() imaging.Format {
	switch x {
	case FormatJpeg:
		return imaging.JPEG
	case FormatPng:
		return imaging.PNG
	case FormatGif:
		return imaging.GIF
	case FormatTiff:
		return imaging.TIFF
	case FormatBmp:
		return imaging.BMP
	default:
		return imaging.JPEG
	}
}

/*
ENUM(
high
medium
low
)
*/
type Quality int

func (x Quality) resampleFilter() imaging.ResampleFilter {
	switch x {
	case QualityHigh:
		return imaging.Lanczos
	case QualityMedium:
		return imaging.Box
	case QualityLow:
		return imaging.NearestNeighbor
	default:
		return imaging.Linear
	}
}

/*
ENUM(
fit
fill
)
*/
type ResizeMode int

func (s *Service) FormatFromExtension(ext string) (Format, error) {
	format, err := imaging.FormatFromExtension(ext)
	if err != nil {
		return -1, ErrUnsupportedFormat
	}
	switch format {
	case imaging.JPEG:
		return FormatJpeg, nil
	case imaging.PNG:
		return FormatPng, nil
	case imaging.GIF:
		return FormatGif, nil
	case imaging.TIFF:
		return FormatTiff, nil
	case imaging.BMP:
		return FormatBmp, nil
	}
	return -1, ErrUnsupportedFormat
}

type resizeConfig struct {
	prioritized bool
	resizeMode  ResizeMode
	quality     Quality
}

type Option func(*resizeConfig)

func WithMode(mode ResizeMode) Option {
	return func(config *resizeConfig) {
		config.resizeMode = mode
	}
}

func WithQuality(quality Quality) Option {
	return func(config *resizeConfig) {
		config.quality = quality
	}
}

func WithHighPriority() Option {
	return func(config *resizeConfig) {
		config.prioritized = true
	}
}

func (s *Service) Resize(ctx context.Context, file afero.File, width, height int, out io.Writer, options ...Option) error {
	config := resizeConfig{
		resizeMode: ResizeModeFit,
		quality:    QualityMedium,
	}
	for _, option := range options {
		option(&config)
	}

	sem := s.lowPrioritySem
	if config.prioritized {
		sem = s.highPrioritySem
	}

	if err := sem.Acquire(ctx, 1); err != nil {
		return err
	}
	defer sem.Release(1)

	format, err := s.FormatFromExtension(filepath.Ext(file.Name()))
	if err != nil {
		return ErrUnsupportedFormat
	}
	img, err := imaging.Decode(file, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	switch config.resizeMode {
	case ResizeModeFill:
		img = imaging.Fill(img, width, height, imaging.Center, config.quality.resampleFilter())
	default:
		img = imaging.Fit(img, width, height, config.quality.resampleFilter())
	}

	return imaging.Encode(out, img, format.toImaging())
}
