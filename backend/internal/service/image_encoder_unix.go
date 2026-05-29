//go:build linux || darwin

package service

import "io"

// defaultEncoder returns the lilliput-based AVIF encoder for Linux/macOS.
func defaultEncoder() func(io.Reader, io.Writer, AVIFConversionSettings) error {
	return encodeAVIFToWriterLilliput
}