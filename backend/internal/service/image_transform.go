package service

import (
	"bytes"
	"fmt"
	"image"
	"io"

	"github.com/gen2brain/avif"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
	_ "image/gif"
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

func avifConversionSettingsFromRuntime(settings RuntimeSettings) AVIFConversionSettings {
	return AVIFConversionSettings{
		Quality: settings.AvifQuality,
		Speed:   settings.AvifSpeed,
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
	img, _, err := image.Decode(source)
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
