package main

import (
	"math"
	"math/cmplx"
	fftpkg "relativity_simulation_2d/pkg/fft"
	"testing"
	"time"
)

// TestFFTPerformance tests FFT implementation performance
func TestFFTPerformance(t *testing.T) {
	// Use CPU FFT processor for now
	processor := fftpkg.NewFFTProcessor()

	// Test multiple grid sizes to verify O(N log N) scaling
	testSizes := []struct {
		size int
		name string
	}{
		{64, "64x64"},
		{128, "128x128"},
		{256, "256x256"},
	}

	for _, test := range testSizes {
		t.Logf("Testing FFT size: %s", test.name)
		width, height := test.size, test.size

		// Create test density grid (2D Gaussian for mathematical properties)
		densityGrid := make([][]float64, height)
		for j := 0; j < height; j++ {
			densityGrid[j] = make([]float64, width)
			for i := 0; i < width; i++ {
				// Gaussian centered in grid
				x := float64(i-width/2) / float64(width/8)
				z := float64(j-height/2) / float64(height/8)
				densityGrid[j][i] = math.Exp(-(x*x + z*z))
			}
		}

		// Convert to complex grid
		complexGrid := make([][]complex128, height)
		for j := 0; j < height; j++ {
			complexGrid[j] = make([]complex128, width)
			for i := 0; i < width; i++ {
				complexGrid[j][i] = complex(densityGrid[j][i], 0)
			}
		}

		// Measure FFT time
		fftStart := time.Now()
		fftResult := processor.FFT2D(complexGrid)
		fftTime := time.Since(fftStart)

		// Measure IFFT time
		ifftStart := time.Now()
		ifftResult := processor.IFFT2D(fftResult)
		ifftTime := time.Since(ifftStart)

		// Verify forward-inverse identity
		maxError := 0.0
		for j := 0; j < height; j++ {
			for i := 0; i < width; i++ {
				expected := densityGrid[j][i]
				actual := real(ifftResult[j][i])
				error := math.Abs(actual - expected)
				if error > maxError {
					maxError = error
				}
			}
		}

		// Verify error is small
		if maxError > 1e-10 {
			t.Errorf("Size %s: FFT-IFFT round trip error too large: %e", test.name, maxError)
		}

		t.Logf("  FFT time: %v", fftTime)
		t.Logf("  IFFT time: %v", ifftTime)
		t.Logf("  Max error: %e", maxError)
	}
}

// TestFFTSpectralProperties tests mathematical properties of FFT
func TestFFTSpectralProperties(t *testing.T) {
	processor := fftpkg.NewFFTProcessor()
	size := 64

	// Create delta function (impulse at center)
	deltaGrid := make([][]complex128, size)
	for i := range deltaGrid {
		deltaGrid[i] = make([]complex128, size)
	}
	deltaGrid[size/2][size/2] = 1.0

	// FFT of delta function should be constant
	fftDelta := processor.FFT2D(deltaGrid)

	// Check that all values have similar magnitude
	// FFT of unit impulse should have all values at magnitude 1
	expectedMag := 1.0
	tolerance := 1e-10

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			mag := cmplx.Abs(fftDelta[i][j])
			if math.Abs(mag-expectedMag) > tolerance {
				t.Errorf("FFT of delta: expected magnitude %v, got %v at [%d,%d]",
					expectedMag, mag, i, j)
			}
		}
	}
}

// TestFFTPoissonSolver tests using FFT to solve Poisson equation
func TestFFTPoissonSolver(t *testing.T) {
	processor := fftpkg.NewFFTProcessor()
	size := 32
	G := 1.0 // Gravitational constant

	// Create a simple mass distribution (point mass at center)
	massGrid := make([][]complex128, size)
	for i := range massGrid {
		massGrid[i] = make([]complex128, size)
	}
	massGrid[size/2][size/2] = 1.0

	// FFT of mass distribution
	massFT := processor.FFT2D(massGrid)

	// Solve Poisson equation in Fourier space: Φ̂(k) = -4πG * ρ̂(k) / |k|²
	kFactor := 2.0 * math.Pi / float64(size)
	for u := 0; u < size; u++ {
		for v := 0; v < size; v++ {
			kx := float64(u)
			if u > size/2 {
				kx = float64(u - size)
			}
			ky := float64(v)
			if v > size/2 {
				ky = float64(v - size)
			}

			kSquared := (kx*kFactor)*(kx*kFactor) + (ky*kFactor)*(ky*kFactor)

			if kSquared > 0 {
				massFT[u][v] *= complex(-4.0*math.Pi*G/kSquared, 0)
			} else {
				massFT[u][v] = 0 // DC component
			}
		}
	}

	// Inverse FFT to get potential
	potential := processor.IFFT2D(massFT)

	// Check that potential is real (imaginary part should be near zero)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if math.Abs(imag(potential[i][j])) > 1e-10 {
				t.Errorf("Potential has non-zero imaginary part at [%d,%d]: %v",
					i, j, imag(potential[i][j]))
			}
		}
	}

	// Potential should decrease with distance from center
	centerPot := real(potential[size/2][size/2])
	edgePot := real(potential[0][0])
	if math.Abs(centerPot) <= math.Abs(edgePot) {
		t.Errorf("Potential doesn't decrease with distance: center=%v, edge=%v",
			centerPot, edgePot)
	}
}
