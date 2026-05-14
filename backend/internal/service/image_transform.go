package service

import (
	"bytes"
	"fmt"
	"image"

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
	img, _, err := image.Decode(bytes.NewReader(payload))
	if err != nil {
		return nil, WithUserMessage(ErrInvalidInput, "file type is not allowed")
	}

	var output bytes.Buffer
	if err := avif.Encode(&output, img, avif.Options{
		Quality: settings.Quality,
		Speed:   settings.Speed,
	}); err != nil {
		return nil, fmt.Errorf("%w: failed to convert image to avif", ErrDependencyUnavailable)
	}

	return output.Bytes(), nil
}
