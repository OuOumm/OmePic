package service

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"
	"strings"
)

const uploadSniffBytes = 512

type verifiedUploadImage struct {
	MIMEType          string
	Format            string
	Width             int
	Height            int
	DetectedContent   string
	DetectedMagicMIME string
}

func verifyUploadImageSource(source preparedUploadSource, requestedMIME string, allowedMIMETypes []string) (verifiedUploadImage, error) {
	reader, err := source.Open()
	if err != nil {
		return verifiedUploadImage{}, fmt.Errorf("%w: failed to open upload source", ErrDependencyUnavailable)
	}
	defer reader.Close()

	head, err := readUploadHeader(reader)
	if err != nil {
		return verifiedUploadImage{}, err
	}
	detectedContent := http.DetectContentType(head)
	detectedMagic := detectImageMagicMIME(head)

	configReader, err := source.Open()
	if err != nil {
		return verifiedUploadImage{}, fmt.Errorf("%w: failed to reopen upload source", ErrDependencyUnavailable)
	}
	defer configReader.Close()

	config, format, err := image.DecodeConfig(configReader)
	if err != nil {
		return verifiedUploadImage{}, WithUserMessage(ErrInvalidInput, "file is not a valid image")
	}
	if config.Width <= 0 || config.Height <= 0 {
		return verifiedUploadImage{}, WithUserMessage(ErrInvalidInput, "image dimensions are invalid")
	}

	realMIME, ok := imageFormatToMIME(format)
	if !ok {
		return verifiedUploadImage{}, WithUserMessage(ErrInvalidInput, "file type is not allowed")
	}
	if detectedMagic != "" && detectedMagic != realMIME {
		return verifiedUploadImage{}, WithUserMessage(ErrInvalidInput, "file content does not match its detected type")
	}
	if !requestMIMECompatible(requestedMIME, realMIME) {
		return verifiedUploadImage{}, WithUserMessage(ErrInvalidInput, "file content type does not match the uploaded image")
	}
	if !allowedMIME(allowedMIMETypes, realMIME) {
		return verifiedUploadImage{}, WithUserMessage(ErrInvalidInput, "file MIME type is not allowed")
	}

	return verifiedUploadImage{
		MIMEType:          realMIME,
		Format:            format,
		Width:             config.Width,
		Height:            config.Height,
		DetectedContent:   detectedContent,
		DetectedMagicMIME: detectedMagic,
	}, nil
}

func readUploadHeader(reader io.Reader) ([]byte, error) {
	buf := make([]byte, uploadSniffBytes)
	n, err := io.ReadFull(reader, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("%w: failed to read upload header", ErrDependencyUnavailable)
	}
	return buf[:n], nil
}

func imageFormatToMIME(format string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "avif":
		return "image/avif", true
	case "bmp":
		return "image/bmp", true
	case "gif":
		return "image/gif", true
	case "jpeg":
		return "image/jpeg", true
	case "png":
		return "image/png", true
	case "webp":
		return "image/webp", true
	default:
		return "", false
	}
}

func requestMIMECompatible(requestedMIME string, realMIME string) bool {
	requested := normalizeUploadMIME(requestedMIME)
	if requested == "" || requested == "application/octet-stream" {
		return true
	}
	return requested == realMIME
}

func normalizeUploadMIME(value string) string {
	mimeType := strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
	if mimeType == "image/jpg" {
		return "image/jpeg"
	}
	return mimeType
}

func detectImageMagicMIME(head []byte) string {
	switch {
	case bytes.HasPrefix(head, []byte("\x89PNG\r\n\x1a\n")):
		return "image/png"
	case len(head) >= 3 && head[0] == 0xff && head[1] == 0xd8 && head[2] == 0xff:
		return "image/jpeg"
	case bytes.HasPrefix(head, []byte("GIF87a")) || bytes.HasPrefix(head, []byte("GIF89a")):
		return "image/gif"
	case len(head) >= 12 && bytes.Equal(head[0:4], []byte("RIFF")) && bytes.Equal(head[8:12], []byte("WEBP")):
		return "image/webp"
	case len(head) >= 2 && head[0] == 'B' && head[1] == 'M':
		return "image/bmp"
	case len(head) >= 12 && bytes.Equal(head[4:8], []byte("ftyp")) && isAVIFBrand(head[8:12]):
		return "image/avif"
	default:
		return ""
	}
}

func isAVIFBrand(brand []byte) bool {
	return bytes.Equal(brand, []byte("avif")) || bytes.Equal(brand, []byte("avis"))
}
