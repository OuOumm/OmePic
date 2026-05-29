package service

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"io"

	"github.com/gen2brain/avif"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
	_ "image/jpeg"
	_ "image/png"
)

const (
	publicImageExtension = ".avif"
	publicImageMIMEType  = "image/avif"
)

type AVIFConversionSettings struct {
	Quality int
	Speed   int
}

func avifConversionSettings(quality int, speed int) AVIFConversionSettings {
	if quality <= 0 {
		quality = DefaultAVIFQuality
	}
	if speed <= 0 {
		speed = DefaultAVIFSpeed
	}
	return AVIFConversionSettings{
		Quality: quality,
		Speed:   speed,
	}
}

func convertToAVIFWithSettings(payload []byte, settings AVIFConversionSettings) ([]byte, error) {
	var output bytes.Buffer
	if err := encodeAVIFToWriter(bytes.NewReader(payload), &output, settings); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func encodeAVIFToWriter(source io.Reader, target io.Writer, settings AVIFConversionSettings) error {
	payload, err := io.ReadAll(source)
	if err != nil {
		return fmt.Errorf("%w: failed to read source data", ErrDependencyUnavailable)
	}
	if isAnimatedGIF(payload) {
		return fmt.Errorf("%w: animated GIF to animated AVIF requires the Linux/macOS lilliput encoder", ErrDependencyUnavailable)
	}

	img, _, err := image.Decode(bytes.NewReader(payload))
	if err != nil {
		return WithUserMessage(ErrInvalidInput, "file type is not allowed")
	}
	if err := avif.Encode(target, img, avif.Options{
		Quality: settings.Quality,
		Speed:   settings.Speed,
	}); err != nil {
		return fmt.Errorf("%w: failed to convert image to avif", ErrDependencyUnavailable)
	}
	return nil
}

func isAnimatedGIF(payload []byte) bool {
	gifImage, err := gif.DecodeAll(bytes.NewReader(payload))
	return err == nil && len(gifImage.Image) > 1
}
