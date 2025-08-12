package physics

import (
	"math"
	"testing"
)

func TestDepositMass(t *testing.T) {
	// Test 2.2: Test Cloud-in-Cell mass deposition

	width := 10
	height := 10

	// Create a single particle at a known position
	particles := []*Particle{
		{
			Position: NewVec3(2.5, 0, 3.5), // Should deposit to cells (2,3), (3,3), (2,4), (3,4)
			Mass:     100.0,
		},
	}

	// Deposit mass to grid
	grid := DepositMassToGrid(particles, width, height)

	// Check that mass is deposited correctly using Cloud-in-Cell
	// fx = 0.5, fz = 0.5, so mass should be split equally among 4 cells
	expectedMassPerCell := 100.0 * 0.5 * 0.5 // 25.0

	// Convert to grid indices (adding offset for centering)
	gx := 2.5 + float64(width)/2.0  // 7.5
	gz := 3.5 + float64(height)/2.0 // 8.5
	i := int(gx)                    // 7
	j := int(gz)                    // 8

	tolerance := 0.001
	if math.Abs(grid[i][j]-expectedMassPerCell) > tolerance {
		t.Errorf("Mass at (%d,%d) incorrect: got %f, expected %f", i, j, grid[i][j], expectedMassPerCell)
	}
	if math.Abs(grid[i+1][j]-expectedMassPerCell) > tolerance {
		t.Errorf("Mass at (%d,%d) incorrect: got %f, expected %f", i+1, j, grid[i+1][j], expectedMassPerCell)
	}
	if math.Abs(grid[i][j+1]-expectedMassPerCell) > tolerance {
		t.Errorf("Mass at (%d,%d) incorrect: got %f, expected %f", i, j+1, grid[i][j+1], expectedMassPerCell)
	}
	if math.Abs(grid[i+1][j+1]-expectedMassPerCell) > tolerance {
		t.Errorf("Mass at (%d,%d) incorrect: got %f, expected %f", i+1, j+1, grid[i+1][j+1], expectedMassPerCell)
	}

	// Check total mass conservation
	totalMass := 0.0
	for i := range grid {
		for j := range grid[i] {
			totalMass += grid[i][j]
		}
	}
	if math.Abs(totalMass-100.0) > tolerance {
		t.Errorf("Total mass not conserved: got %f, expected 100.0", totalMass)
	}
}

func TestSolvePoissonEquation(t *testing.T) {
	// Test solving Poisson equation ∇²Φ = 4πGρ

	width := 32 // Use power of 2 for FFT
	height := 32
	gravitationalConstant := 1.0

	// Create a simple mass distribution (point mass at center)
	massGrid := make([][]float64, width)
	for i := range massGrid {
		massGrid[i] = make([]float64, height)
	}
	massGrid[width/2][height/2] = 100.0

	// Solve for potential
	potentialGrid := SolvePoissonFFT(massGrid, width, height, gravitationalConstant)

	// Check that potential is negative (attractive)
	if potentialGrid[width/2][height/2] >= 0 {
		t.Error("Potential at mass location should be negative (attractive)")
	}

	// Check that potential decreases with distance (in magnitude)
	centerPotential := math.Abs(potentialGrid[width/2][height/2])
	nearPotential := math.Abs(potentialGrid[width/2+1][height/2])
	farPotential := math.Abs(potentialGrid[width/2+5][height/2])

	if nearPotential >= centerPotential {
		t.Error("Potential magnitude should decrease with distance")
	}
	if farPotential >= nearPotential {
		t.Error("Potential magnitude should continue decreasing with distance")
	}
}

func TestCalculateGradient(t *testing.T) {
	// Test gradient calculation a = -∇Φ

	width := 10
	height := 10

	// Create a simple linear potential for easy verification
	potentialGrid := make([][]float64, width)
	for i := range potentialGrid {
		potentialGrid[i] = make([]float64, height)
		for j := range potentialGrid[i] {
			// Linear potential: Φ = x + 2*z
			potentialGrid[i][j] = float64(i) + 2.0*float64(j)
		}
	}

	// Calculate gradient
	forceField := CalculateGradient(potentialGrid, width, height)

	// For linear potential, gradient should be constant
	// ∂Φ/∂x = 1, so ax = -1
	// ∂Φ/∂z = 2, so az = -2
	expectedAx := -1.0 // Central difference gives (next - prev) / 2 = (i+1 - (i-1)) / 2 = 2/2 = 1, negated = -1
	expectedAz := -2.0 // Central difference gives (j+1 - (j-1)) / 2 = 4/2 = 2, negated = -2

	tolerance := 0.01
	// Check interior points (avoiding boundaries)
	for i := 1; i < width-1; i++ {
		for j := 1; j < height-1; j++ {
			if math.Abs(forceField.AccelFieldX[i][j]-expectedAx) > tolerance {
				t.Errorf("AccelX at (%d,%d) incorrect: got %f, expected %f",
					i, j, forceField.AccelFieldX[i][j], expectedAx)
			}
			if math.Abs(forceField.AccelFieldZ[i][j]-expectedAz) > tolerance {
				t.Errorf("AccelZ at (%d,%d) incorrect: got %f, expected %f",
					i, j, forceField.AccelFieldZ[i][j], expectedAz)
			}
		}
	}
}

func TestInterpolateAcceleration(t *testing.T) {
	// Test bilinear interpolation of acceleration field to particle position

	width := 10
	height := 10

	// Create a simple uniform acceleration field
	forceField := &ForceField{
		AccelFieldX: make([][]float64, width),
		AccelFieldZ: make([][]float64, height),
		Width:       width,
		Height:      height,
	}

	for i := range forceField.AccelFieldX {
		forceField.AccelFieldX[i] = make([]float64, height)
		forceField.AccelFieldZ[i] = make([]float64, height)
		for j := range forceField.AccelFieldX[i] {
			forceField.AccelFieldX[i][j] = -1.0 // Uniform acceleration in X
			forceField.AccelFieldZ[i][j] = -2.0 // Uniform acceleration in Z
		}
	}

	// Test interpolation at various positions
	testCases := []struct {
		position   Vec3
		expectedAx float64
		expectedAz float64
	}{
		{NewVec3(0, 0, 0), -1.0, -2.0},     // Center of cell
		{NewVec3(0.5, 0, 0.5), -1.0, -2.0}, // Between cells (uniform field)
		{NewVec3(-2, 0, -2), -1.0, -2.0},   // Another position
	}

	tolerance := 0.001
	for _, tc := range testCases {
		ax, az := InterpolateAcceleration(tc.position, forceField)
		if math.Abs(ax-tc.expectedAx) > tolerance {
			t.Errorf("Interpolated ax at %v incorrect: got %f, expected %f",
				tc.position, ax, tc.expectedAx)
		}
		if math.Abs(az-tc.expectedAz) > tolerance {
			t.Errorf("Interpolated az at %v incorrect: got %f, expected %f",
				tc.position, az, tc.expectedAz)
		}
	}
}

func TestFullForceCalculationPipeline(t *testing.T) {
	// Test the complete force calculation pipeline

	width := 32
	height := 32
	gravitationalConstant := 1.0

	// Create two particles that should attract each other
	particles := []*Particle{
		{
			Position: NewVec3(-5, 0, 0),
			Velocity: NewVec3(0, 0, 0),
			Mass:     100.0,
		},
		{
			Position: NewVec3(5, 0, 0),
			Velocity: NewVec3(0, 0, 0),
			Mass:     100.0,
		},
	}

	// Step 1: Deposit mass
	massGrid := DepositMassToGrid(particles, width, height)

	// Step 2: Solve Poisson equation
	potentialGrid := SolvePoissonFFT(massGrid, width, height, gravitationalConstant)

	// Step 3: Calculate gradient
	forceField := CalculateGradient(potentialGrid, width, height)

	// Step 4: Interpolate forces to particles
	ax1, _ := InterpolateAcceleration(particles[0].Position, forceField)
	ax2, _ := InterpolateAcceleration(particles[1].Position, forceField)

	// Particles should experience opposite forces (attraction)
	if ax1*ax2 >= 0 {
		t.Error("Particles should experience opposite X accelerations (attraction)")
	}

	// First particle should be pulled right (positive ax)
	if ax1 <= 0 {
		t.Error("Left particle should be pulled right (positive acceleration)")
	}

	// Second particle should be pulled left (negative ax)
	if ax2 >= 0 {
		t.Error("Right particle should be pulled left (negative acceleration)")
	}
}
