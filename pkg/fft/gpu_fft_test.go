package fft

import (
	"math"
	"math/cmplx"
	"testing"
	"time"
)

// TestGPUFFTProcessor tests the GPU FFT implementation
func TestGPUFFTProcessor(t *testing.T) {
	// Try to create GPU processor
	processor, err := NewGPUFFTProcessor()
	if err != nil {
		t.Skip("GPU not available:", err)
	}
	defer processor.(*GPUFFTProcessor).Cleanup()

	// Test that it implements the interface
	var _ = processor
}

// TestGPUFFTRoundTrip tests GPU FFT forward-inverse identity
func TestGPUFFTRoundTrip(t *testing.T) {
	processor, err := NewGPUFFTProcessor()
	if err != nil {
		t.Skip("GPU not available:", err)
	}
	defer processor.(*GPUFFTProcessor).Cleanup()

	// Test 2D FFT round-trip
	size := 32
	input := make([][]complex128, size)
	for i := range input {
		input[i] = make([]complex128, size)
		for j := range input[i] {
			// Create a test pattern
			x := float64(i-size/2) / float64(size/4)
			y := float64(j-size/2) / float64(size/4)
			input[i][j] = complex(math.Exp(-(x*x + y*y)), 0)
		}
	}

	// Forward FFT
	fftResult := processor.FFT2D(input)

	// Inverse FFT
	ifftResult := processor.IFFT2D(fftResult)

	// Check round-trip accuracy
	maxError := 0.0
	for i := range input {
		for j := range input[i] {
			error := cmplx.Abs(ifftResult[i][j] - input[i][j])
			if error > maxError {
				maxError = error
			}
		}
	}

	if maxError > 1e-10 {
		t.Errorf("GPU FFT round-trip error too large: %e", maxError)
	}
}

// TestGPUvsCPUFFT compares GPU and CPU FFT results
func TestGPUvsCPUFFT(t *testing.T) {
	gpuProcessor, err := NewGPUFFTProcessor()
	if err != nil {
		t.Skip("GPU not available:", err)
	}
	defer gpuProcessor.(*GPUFFTProcessor).Cleanup()

	cpuProcessor := NewFFTProcessor()

	// Create test data
	size := 16
	input := make([][]complex128, size)
	for i := range input {
		input[i] = make([]complex128, size)
		for j := range input[i] {
			input[i][j] = complex(float64(i+j), 0)
		}
	}

	// Compute FFT with both processors
	cpuResult := cpuProcessor.FFT2D(input)
	gpuResult := gpuProcessor.FFT2D(input)

	// Compare results
	maxDiff := 0.0
	for i := range cpuResult {
		for j := range cpuResult[i] {
			diff := cmplx.Abs(cpuResult[i][j] - gpuResult[i][j])
			if diff > maxDiff {
				maxDiff = diff
			}
		}
	}

	// GPU and CPU should give nearly identical results
	if maxDiff > 1e-6 {
		t.Errorf("GPU vs CPU FFT difference too large: %e", maxDiff)
	}
}

// TestGPUFFTPerformance compares GPU vs CPU performance
func TestGPUFFTPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	gpuProcessor, err := NewGPUFFTProcessor()
	if err != nil {
		t.Skip("GPU not available:", err)
	}
	defer gpuProcessor.(*GPUFFTProcessor).Cleanup()

	cpuProcessor := NewFFTProcessor()

	// Test with a larger grid for meaningful performance comparison
	size := 256
	input := make([][]complex128, size)
	for i := range input {
		input[i] = make([]complex128, size)
		for j := range input[i] {
			input[i][j] = complex(math.Sin(float64(i)*0.1)*math.Cos(float64(j)*0.1), 0)
		}
	}

	// Warm up
	_ = cpuProcessor.FFT2D(input)
	_ = gpuProcessor.FFT2D(input)

	// Benchmark CPU
	cpuIterations := 10
	cpuStart := time.Now()
	for i := 0; i < cpuIterations; i++ {
		_ = cpuProcessor.FFT2D(input)
	}
	cpuTime := time.Since(cpuStart)

	// Benchmark GPU
	gpuIterations := 10
	gpuStart := time.Now()
	for i := 0; i < gpuIterations; i++ {
		_ = gpuProcessor.FFT2D(input)
	}
	gpuTime := time.Since(gpuStart)

	t.Logf("CPU FFT time: %v", cpuTime)
	t.Logf("GPU FFT time: %v", gpuTime)
	t.Logf("Speedup: %.2fx", float64(cpuTime)/float64(gpuTime))

	// GPU should be faster for large grids
	if gpuTime > cpuTime {
		t.Log("Warning: GPU FFT slower than CPU - may indicate driver or hardware issues")
	}
}
