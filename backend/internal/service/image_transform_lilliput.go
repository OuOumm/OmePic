//go:build linux || darwin

package service

import (
	"errors"
	"fmt"
	"io"

	"github.com/discord/lilliput"
)

const (
	// maxAnimationFrames is the maximum number of frames allowed for
	// animated images. Exceeding this limit will reject the upload.
	maxAnimationFrames = 300

	// avifEncodeQuality and avifEncodeSpeed are the keys passed to
	// lilliput's EncodeOptions map, matching the C++ enum values:
	//   AVIF_QUALITY = 1, AVIF_SPEED = 2
	avifEncodeQuality = 1
	avifEncodeSpeed   = 2
)

// ErrTooManyFrames is returned when an animated image exceeds the
// maximum allowed frame count.
var ErrTooManyFrames = errors.New("too many animation frames")

// encodeAVIFToWriterLilliput converts an image from source to AVIF format
// and writes the result to target. It preserves animation frames for GIF,
// WebP, and AVIF inputs.
//
// Signature matches: func(io.Reader, io.Writer, AVIFConversionSettings) error
func encodeAVIFToWriterLilliput(source io.Reader, target io.Writer, settings AVIFConversionSettings) error {
	data, err := io.ReadAll(source)
	if err != nil {
		return fmt.Errorf("%w: failed to read source data", ErrDependencyUnavailable)
	}
	if len(data) == 0 {
		return WithUserMessage(ErrInvalidInput, "file type is not allowed")
	}

	// Step 1: check frame count (consumes a decoder; do NOT reuse for Transform)
	animatedInput, err := checkFrameCount(data)
	if err != nil {
		return err
	}

	// Step 2: create a fresh decoder for actual conversion
	decoder, err := lilliput.NewDecoder(data)
	if err != nil {
		return WithUserMessage(ErrInvalidInput, "file type is not allowed")
	}
	defer decoder.Close()

	// Step 3: set up ImageOps for no-resize conversion
	ops := lilliput.NewImageOps(0)
	defer ops.Close()

	opts := &lilliput.ImageOptions{
		FileType:             ".avif",
		Width:                0,
		Height:               0,
		ResizeMethod:         lilliput.ImageOpsNoResize,
		NormalizeOrientation: true,
		EncodeOptions: map[int]int{
			avifEncodeQuality: settings.Quality,
			avifEncodeSpeed:   settings.Speed,
		},
		MaxEncodeFrames:       maxAnimationFrames,
		DisableAnimatedOutput: false,
	}

	// Step 4: transform and write output
	// Allocate output buffer (estimated ≤ 2× source size)
	buf := make([]byte, len(data)*2)
	result, err := ops.Transform(decoder, opts, buf)
	if err != nil {
		if errors.Is(err, lilliput.ErrBufTooSmall) {
			// Retry with larger buffer
			buf = make([]byte, len(data)*4)
			result, err = ops.Transform(decoder, opts, buf)
		}
		if err != nil {
			return fmt.Errorf("%w: failed to convert image to avif", ErrDependencyUnavailable)
		}
	}

	if animatedInput {
		animatedOutput, err := isLilliputAnimated(result)
		if err != nil {
			return fmt.Errorf("%w: failed to validate animated avif output", ErrDependencyUnavailable)
		}
		if !animatedOutput {
			return fmt.Errorf("%w: animated image conversion produced a static avif", ErrDependencyUnavailable)
		}
	}

	if _, err := target.Write(result); err != nil {
		return fmt.Errorf("%w: failed to write converted image", ErrDependencyUnavailable)
	}
	return nil
}

// checkFrameCount creates a decoder from data, inspects the image header,
// and if animated, counts the frames. Returns true when the source image is
// animated. Returns ErrTooManyFrames if the frame count exceeds
// maxAnimationFrames.
func checkFrameCount(data []byte) (bool, error) {
	decoder, err := lilliput.NewDecoder(data)
	if err != nil {
		return false, WithUserMessage(ErrInvalidInput, "file type is not allowed")
	}
	defer decoder.Close()

	header, err := decoder.Header()
	if err != nil {
		return false, WithUserMessage(ErrInvalidInput, "file type is not allowed")
	}

	// Static image — no need to count
	if !header.IsAnimated() {
		return false, nil
	}

	// Count frames by decoding each one
	fb := lilliput.NewFramebuffer(header.Width(), header.Height())
	defer fb.Close()

	count := 0
	for count <= maxAnimationFrames {
		err := decoder.DecodeTo(fb)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return true, fmt.Errorf("%w: failed to decode animation frame", ErrDependencyUnavailable)
		}
		count++
	}

	if count > maxAnimationFrames {
		return true, WithUserMessage(ErrTooManyFrames,
			fmt.Sprintf("animated image has too many frames (max %d)", maxAnimationFrames))
	}
	return true, nil
}

func isLilliputAnimated(data []byte) (bool, error) {
	decoder, err := lilliput.NewDecoder(data)
	if err != nil {
		return false, err
	}
	defer decoder.Close()

	header, err := decoder.Header()
	if err != nil {
		return false, err
	}
	return header.IsAnimated(), nil
}
