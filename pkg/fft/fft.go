package fft

import (
	"github.com/mjibson/go-dsp/fft"
)

// FFTProcessor defines the interface for FFT operations
type FFTProcessor interface {
	FFT1D(input []complex128) []complex128
	IFFT1D(input []complex128) []complex128
	FFT2D(input [][]complex128) [][]complex128
	IFFT2D(input [][]complex128) [][]complex128
}

// CPUFFTProcessor implements FFT operations using CPU
type CPUFFTProcessor struct{}

// NewFFTProcessor creates a new FFT processor
func NewFFTProcessor() FFTProcessor {
	return &CPUFFTProcessor{}
}

// FFT1D performs one-dimensional FFT
func (p *CPUFFTProcessor) FFT1D(input []complex128) []complex128 {
	return fft.FFT(input)
}

// IFFT1D performs one-dimensional inverse FFT
func (p *CPUFFTProcessor) IFFT1D(input []complex128) []complex128 {
	return fft.IFFT(input)
}

// FFT2D performs two-dimensional FFT
func (p *CPUFFTProcessor) FFT2D(input [][]complex128) [][]complex128 {
	return fft.FFT2(input)
}

// IFFT2D performs two-dimensional inverse FFT
func (p *CPUFFTProcessor) IFFT2D(input [][]complex128) [][]complex128 {
	return fft.IFFT2(input)
}

// FFT2Real performs 2D FFT on real-valued input and returns real-valued output
// This is a convenience function for physics simulations
func FFT2Real(input [][]float64) [][]complex128 {
	width := len(input)
	if width == 0 {
		return nil
	}
	height := len(input[0])

	// Convert to complex
	complexGrid := make([][]complex128, width)
	for i := range complexGrid {
		complexGrid[i] = make([]complex128, height)
		for j := range complexGrid[i] {
			complexGrid[i][j] = complex(input[i][j], 0)
		}
	}

	// Perform FFT
	processor := NewFFTProcessor()
	return processor.FFT2D(complexGrid)
}

// IFFT2Real performs 2D inverse FFT and returns only the real part
// This is a convenience function for physics simulations
func IFFT2Real(input [][]complex128) [][]float64 {
	processor := NewFFTProcessor()
	result := processor.IFFT2D(input)

	width := len(result)
	if width == 0 {
		return nil
	}
	height := len(result[0])

	// Extract real part
	realGrid := make([][]float64, width)
	for i := range realGrid {
		realGrid[i] = make([]float64, height)
		for j := range realGrid[i] {
			realGrid[i][j] = real(result[i][j])
		}
	}

	return realGrid
}
