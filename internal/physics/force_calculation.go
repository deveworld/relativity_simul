package physics

import (
	"math"
	"relativity_simulation_2d/pkg/fft"
)

// ForceField represents the gravitational acceleration field
type ForceField struct {
	AccelFieldX [][]float64
	AccelFieldZ [][]float64
	Width       int
	Height      int
}

// DepositMassToGrid distributes particle mass to grid using Cloud-in-Cell
func DepositMassToGrid(particles []*Particle, width, height int) [][]float64 {
	// Initialize mass density grid
	grid := make([][]float64, width)
	for i := range grid {
		grid[i] = make([]float64, height)
	}

	// Deposit each particle's mass
	for _, p := range particles {
		// Find grid cell coordinates and fractional parts
		gx := p.Position.X + float64(width)/2.0
		gz := p.Position.Z + float64(height)/2.0
		i := int(gx)
		j := int(gz)
		fx := gx - float64(i)
		fz := gz - float64(j)

		// Distribute mass to 4 nearest cells (Cloud-in-Cell)
		if i >= 0 && i < width-1 && j >= 0 && j < height-1 {
			grid[i][j] += float64(p.Mass) * (1 - fx) * (1 - fz)
			grid[i+1][j] += float64(p.Mass) * fx * (1 - fz)
			grid[i][j+1] += float64(p.Mass) * (1 - fx) * fz
			grid[i+1][j+1] += float64(p.Mass) * fx * fz
		}
	}

	return grid
}

// SolvePoissonFFT solves ∇²Φ = 4πGρ using FFT
func SolvePoissonFFT(massGrid [][]float64, width, height int, gravitationalConstant float64) [][]float64 {
	// Convert mass density grid to complex numbers for FFT
	complexGrid := make([][]complex128, width)
	for i := range complexGrid {
		complexGrid[i] = make([]complex128, height)
		for j := range complexGrid[i] {
			complexGrid[i][j] = complex(massGrid[i][j], 0)
		}
	}

	// Create FFT processor
	processor := fft.NewFFTProcessor()

	// 2D FFT of the mass density
	fftGrid := processor.FFT2D(complexGrid)

	// Solve in Fourier space: Φ̂(k) = -4πG * ρ̂(k) / |k|²
	kxFactor := 2.0 * math.Pi / float64(width)
	kzFactor := 2.0 * math.Pi / float64(height)

	for u := 0; u < width; u++ {
		for v := 0; v < height; v++ {
			// Calculate wave vector k
			kx := float64(u)
			if u > width/2 {
				kx = float64(u - width)
			}
			kz := float64(v)
			if v > height/2 {
				kz = float64(v - height)
			}

			kSquared := (kx*kxFactor)*(kx*kxFactor) + (kz*kzFactor)*(kz*kzFactor)

			if kSquared == 0 {
				fftGrid[u][v] = 0 // Ignore the DC component (average potential)
			} else {
				// Standard gravitational Poisson equation: ∇²Φ = 4πGρ
				scalingFactor := -4.0 * math.Pi * gravitationalConstant / kSquared
				fftGrid[u][v] *= complex(scalingFactor, 0)
			}
		}
	}

	// Inverse 2D FFT to get the potential grid in real space
	potentialComplex := processor.IFFT2D(fftGrid)

	// Copy real part to potential grid
	potentialGrid := make([][]float64, width)
	for i := range potentialGrid {
		potentialGrid[i] = make([]float64, height)
		for j := range potentialGrid[i] {
			potentialGrid[i][j] = real(potentialComplex[i][j])
		}
	}

	return potentialGrid
}

// CalculateGradient computes acceleration a = -∇Φ using central differences
func CalculateGradient(potentialGrid [][]float64, width, height int) *ForceField {
	forceField := &ForceField{
		AccelFieldX: make([][]float64, width),
		AccelFieldZ: make([][]float64, height),
		Width:       width,
		Height:      height,
	}

	for i := range forceField.AccelFieldX {
		forceField.AccelFieldX[i] = make([]float64, height)
		forceField.AccelFieldZ[i] = make([]float64, height)
	}

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			// Use modulo arithmetic for periodic (wrapping) boundaries
			prevI := (i - 1 + width) % width
			nextI := (i + 1) % width
			prevJ := (j - 1 + height) % height
			nextJ := (j + 1) % height

			// Central difference for gradient with periodic boundaries
			forceField.AccelFieldX[i][j] = -(potentialGrid[nextI][j] - potentialGrid[prevI][j]) / 2.0
			forceField.AccelFieldZ[i][j] = -(potentialGrid[i][nextJ] - potentialGrid[i][prevJ]) / 2.0
		}
	}

	return forceField
}

// InterpolateAcceleration interpolates acceleration from grid to particle position
func InterpolateAcceleration(position Vec3, forceField *ForceField) (ax, az float64) {
	// Find grid cell coordinates and fractional parts for interpolation
	gx := position.X + float64(forceField.Width)/2.0
	gz := position.Z + float64(forceField.Height)/2.0
	i := int(gx)
	j := int(gz)
	fx := gx - float64(i)
	fz := gz - float64(j)

	// Bilinear interpolation
	if i >= 0 && i < forceField.Width-1 && j >= 0 && j < forceField.Height-1 {
		ax1 := forceField.AccelFieldX[i][j]*(1-fz) + forceField.AccelFieldX[i][j+1]*fz
		ax2 := forceField.AccelFieldX[i+1][j]*(1-fz) + forceField.AccelFieldX[i+1][j+1]*fz
		ax = ax1*(1-fx) + ax2*fx

		az1 := forceField.AccelFieldZ[i][j]*(1-fz) + forceField.AccelFieldZ[i][j+1]*fz
		az2 := forceField.AccelFieldZ[i+1][j]*(1-fz) + forceField.AccelFieldZ[i+1][j+1]*fz
		az = az1*(1-fx) + az2*fx
	}

	return ax, az
}

// UpdateVelocities updates particle velocities based on acceleration field (Kick step)
func UpdateVelocities(particles []*Particle, forceField *ForceField, dt float32, forceCorrectionFactor float32) {
	for _, p := range particles {
		ax, az := InterpolateAcceleration(p.Position, forceField)

		// Apply forces with correction factor to approximately remove self-interaction
		p.Velocity.X += ax * float64(dt) * float64(forceCorrectionFactor)
		p.Velocity.Z += az * float64(dt) * float64(forceCorrectionFactor)
	}
}

// UpdatePositions updates the positions of all particles (Drift step)
func UpdatePositions(particles []*Particle, dt float32, width, height int) {
	for _, p := range particles {
		p.Position.X += p.Velocity.X * float64(dt)
		p.Position.Z += p.Velocity.Z * float64(dt)

		// Boundary conditions - wrap around
		if p.Position.X > float64(width)/2.0 {
			p.Position.X = -float64(width) / 2.0
		}
		if p.Position.X < -float64(width)/2.0 {
			p.Position.X = float64(width) / 2.0
		}
		if p.Position.Z > float64(height)/2.0 {
			p.Position.Z = -float64(height) / 2.0
		}
		if p.Position.Z < -float64(height)/2.0 {
			p.Position.Z = float64(height) / 2.0
		}
	}
}
