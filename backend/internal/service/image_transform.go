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

func convertToAVIF(payload []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("%w: file type is not allowed", ErrInvalidInput)
	}

	var output bytes.Buffer
	if err := avif.Encode(&output, img, avif.Options{
		Quality: 60,
		Speed:   8,
	}); err != nil {
		return nil, fmt.Errorf("%w: failed to convert image to avif", ErrDependencyUnavailable)
	}

	return output.Bytes(), nil
}
