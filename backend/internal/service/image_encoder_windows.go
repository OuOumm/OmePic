//go:build !linux && !darwin

package service

import "io"

// defaultEncoder returns the gen2brain/avif encoder for Windows.
func defaultEncoder() func(io.Reader, io.Writer, AVIFConversionSettings) error {
	return encodeAVIFToWriter
}
