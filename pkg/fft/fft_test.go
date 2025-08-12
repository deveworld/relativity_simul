package fft

import (
	"math"
	"math/cmplx"
	"testing"
)

// TestFFTInterface tests the basic FFT interface
func TestFFTInterface(t *testing.T) {
	// Create an FFT processor
	processor := NewFFTProcessor()

	// Test that processor implements the interface
	var _ = processor
}

// TestFFT1D tests one-dimensional FFT
func TestFFT1D(t *testing.T) {
	processor := NewFFTProcessor()

	// Test with a simple signal: [1, 0, 0, 0]
	input := []complex128{1, 0, 0, 0}
	expected := []complex128{1, 1, 1, 1} // FFT of impulse

	result := processor.FFT1D(input)

	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}

	for i := range result {
		if !complexApproxEqual(result[i], expected[i], 1e-10) {
			t.Errorf("Index %d: expected %v, got %v", i, expected[i], result[i])
		}
	}
}

// TestIFFT1D tests one-dimensional inverse FFT
func TestIFFT1D(t *testing.T) {
	processor := NewFFTProcessor()

	// Test that IFFT(FFT(x)) = x
	input := []complex128{1, 2, 3, 4}

	fftResult := processor.FFT1D(input)
	ifftResult := processor.IFFT1D(fftResult)

	if len(ifftResult) != len(input) {
		t.Fatalf("Expected length %d, got %d", len(input), len(ifftResult))
	}

	for i := range ifftResult {
		if !complexApproxEqual(ifftResult[i], input[i], 1e-10) {
			t.Errorf("Index %d: expected %v, got %v", i, input[i], ifftResult[i])
		}
	}
}

// TestFFT2D tests two-dimensional FFT
func TestFFT2D(t *testing.T) {
	processor := NewFFTProcessor()

	// Create a 2x2 test grid
	input := [][]complex128{
		{1, 0},
		{0, 0},
	}

	result := processor.FFT2D(input)

	// Check dimensions
	if len(result) != 2 || len(result[0]) != 2 {
		t.Fatalf("Expected 2x2 grid, got %dx%d", len(result), len(result[0]))
	}

	// Check DC component (sum of all elements)
	dcComponent := result[0][0]
	expectedDC := complex(1, 0)
	if !complexApproxEqual(dcComponent, expectedDC, 1e-10) {
		t.Errorf("DC component: expected %v, got %v", expectedDC, dcComponent)
	}
}

// TestIFFT2D tests two-dimensional inverse FFT
func TestIFFT2D(t *testing.T) {
	processor := NewFFTProcessor()

	// Test that IFFT2D(FFT2D(x)) = x
	input := [][]complex128{
		{1, 2},
		{3, 4},
	}

	fftResult := processor.FFT2D(input)
	ifftResult := processor.IFFT2D(fftResult)

	// Check dimensions
	if len(ifftResult) != len(input) || len(ifftResult[0]) != len(input[0]) {
		t.Fatalf("Dimension mismatch")
	}

	// Check values
	for i := range ifftResult {
		for j := range ifftResult[i] {
			if !complexApproxEqual(ifftResult[i][j], input[i][j], 1e-10) {
				t.Errorf("Position [%d][%d]: expected %v, got %v",
					i, j, input[i][j], ifftResult[i][j])
			}
		}
	}
}

// TestParseval tests Parseval's theorem: sum(|x|^2) = sum(|X|^2)/N
func TestParseval(t *testing.T) {
	processor := NewFFTProcessor()

	input := []complex128{1, 2, 3, 4}

	// Calculate energy in time domain
	timeEnergy := 0.0
	for _, v := range input {
		timeEnergy += real(v * cmplx.Conj(v))
	}

	// Calculate energy in frequency domain
	fftResult := processor.FFT1D(input)
	freqEnergy := 0.0
	for _, v := range fftResult {
		freqEnergy += real(v * cmplx.Conj(v))
	}
	freqEnergy /= float64(len(input))

	// Check Parseval's theorem
	if math.Abs(timeEnergy-freqEnergy) > 1e-10 {
		t.Errorf("Parseval's theorem violated: time=%v, freq=%v",
			timeEnergy, freqEnergy)
	}
}

// TestFFTWithRealSignal tests FFT with real-valued input
func TestFFTWithRealSignal(t *testing.T) {
	processor := NewFFTProcessor()

	// Create a simple cosine wave
	n := 8
	input := make([]complex128, n)
	for i := 0; i < n; i++ {
		input[i] = complex(math.Cos(2*math.Pi*float64(i)/float64(n)), 0)
	}

	result := processor.FFT1D(input)

	// For a cosine with frequency 1, we expect peaks at indices 1 and n-1
	// All other components should be near zero
	for i := range result {
		magnitude := cmplx.Abs(result[i])
		if i == 1 || i == n-1 {
			// Should have significant magnitude
			if magnitude < 3.9 {
				t.Errorf("Expected peak at index %d, got magnitude %v", i, magnitude)
			}
		} else {
			// Should be near zero
			if magnitude > 0.1 {
				t.Errorf("Expected near-zero at index %d, got magnitude %v", i, magnitude)
			}
		}
	}
}

// Helper function to compare complex numbers with tolerance
func complexApproxEqual(a, b complex128, tolerance float64) bool {
	return cmplx.Abs(a-b) < tolerance
}
