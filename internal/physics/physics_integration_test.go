package physics

import (
	"math"
	"testing"
)

// TestPhysicsEngineIntegration verifies the complete physics pipeline
// including particle initialization, force calculation, and time evolution
func TestPhysicsEngineIntegration(t *testing.T) {
	// Setup: Create a small system of particles
	numParticles := 10
	width := 256
	height := 256
	simulationWidth := 100.0
	simulationHeight := 100.0
	dt := float32(0.01)
	gravitationalConstant := 0.1 // Weaker gravity for more stable simulation
	numSteps := 50               // Shorter simulation to reduce numerical drift

	// Initialize particles in a controlled way for testing
	particles := InitializeParticles(numParticles, simulationWidth, simulationHeight)

	// Record initial state for conservation law checks
	initialMomentum := calculateTotalMomentum(particles)

	// Run the physics simulation for multiple steps using RunTimeEvolution
	for step := 0; step < numSteps; step++ {
		// Use the complete physics pipeline function
		forceField := RunTimeEvolution(particles, dt, width, height, gravitationalConstant)
		if forceField == nil {
			t.Fatal("RunTimeEvolution returned nil force field")
		}
	}

	// Verify conservation laws
	finalKE := calculateKineticEnergy(particles)
	finalMomentum := calculateTotalMomentum(particles)

	// For Particle-Mesh (PM) methods with force correction factors,
	// perfect energy conservation is not expected. We just check that
	// energy doesn't explode (similar to existing tests)

	// Check that kinetic energy remains bounded
	// For a system starting at rest with weak gravity, KE should remain small
	maxExpectedKE := float64(numParticles) * 1000.0 // Allow reasonable KE per particle
	if finalKE > maxExpectedKE {
		t.Errorf("Energy exploded: final KE=%v, max expected=%v", finalKE, maxExpectedKE)
	}

	// Momentum should be conserved (no external forces)
	momentumDrift := finalMomentum.Sub(initialMomentum).Length()
	if momentumDrift > 1.0 { // Allow some drift due to numerical errors
		t.Errorf("Momentum conservation violated: initial=%v, final=%v, drift=%v",
			initialMomentum, finalMomentum, momentumDrift)
	}

	// Verify particles remain within bounds (centered coordinate system)
	halfWidth := float64(width) / 2.0
	halfHeight := float64(height) / 2.0
	for i, p := range particles {
		if p.Position.X < -halfWidth || p.Position.X > halfWidth ||
			p.Position.Z < -halfHeight || p.Position.Z > halfHeight {
			t.Errorf("Particle %d out of bounds: position=%v", i, p.Position)
		}
	}

	// Verify no NaN or Inf values
	for i, p := range particles {
		if math.IsNaN(float64(p.Position.X)) || math.IsInf(float64(p.Position.X), 0) ||
			math.IsNaN(float64(p.Position.Y)) || math.IsInf(float64(p.Position.Y), 0) ||
			math.IsNaN(float64(p.Position.Z)) || math.IsInf(float64(p.Position.Z), 0) {
			t.Errorf("Particle %d has invalid position: %v", i, p.Position)
		}
		if math.IsNaN(float64(p.Velocity.X)) || math.IsInf(float64(p.Velocity.X), 0) ||
			math.IsNaN(float64(p.Velocity.Y)) || math.IsInf(float64(p.Velocity.Y), 0) ||
			math.IsNaN(float64(p.Velocity.Z)) || math.IsInf(float64(p.Velocity.Z), 0) {
			t.Errorf("Particle %d has invalid velocity: %v", i, p.Velocity)
		}
	}
}

// TestPhysicsEngineWithCentralMass tests the physics engine with a central massive object
func TestPhysicsEngineWithCentralMass(t *testing.T) {
	// Setup: System with central mass and orbiting particles
	numParticles := 5
	width := 256
	height := 256
	simulationWidth := 100.0
	simulationHeight := 100.0
	centralMass := 1000.0
	dt := float32(0.001)
	gravitationalConstant := 1.0
	numSteps := 1000

	// Initialize with central mass
	particles := InitializeParticlesWithCentralMass(numParticles, simulationWidth, simulationHeight, centralMass)

	// Track orbital parameters
	initialDistances := make([]float64, len(particles))
	for i, p := range particles {
		if i == 0 {
			continue // Skip central mass
		}
		initialDistances[i] = p.Position.Sub(particles[0].Position).Length()
	}

	// Run simulation
	for step := 0; step < numSteps; step++ {
		forceField := RunTimeEvolution(particles, dt, width, height, gravitationalConstant)
		if forceField == nil {
			t.Fatal("RunTimeEvolution returned nil force field")
		}
	}

	// Verify orbits remain stable (particles don't escape or crash)
	for i, p := range particles {
		if i == 0 {
			continue // Skip central mass
		}
		currentDistance := p.Position.Sub(particles[0].Position).Length()

		// Check that particles haven't escaped (distance shouldn't increase too much)
		if currentDistance > initialDistances[i]*2 {
			t.Errorf("Particle %d escaped: initial distance=%v, current=%v",
				i, initialDistances[i], currentDistance)
		}

		// Check that particles haven't crashed into central mass
		if currentDistance < 1.0 {
			t.Errorf("Particle %d crashed into central mass: distance=%v", i, currentDistance)
		}
	}
}

// TestLeapfrogIntegration tests the leapfrog integration method
func TestLeapfrogIntegration(t *testing.T) {
	// Create a simple two-body system
	particles := []*Particle{
		NewParticle(1.0, 0, 0, 0, 0, 0, 1),
		NewParticle(1.0, 10, 0, 0, 0, 0, -1),
	}

	width := 256
	height := 256
	dt := float32(0.01)
	gravitationalConstant := 1.0
	numSteps := 100

	// Record initial center of mass
	initialCOM := calculateCenterOfMass(particles)

	// Run leapfrog integration
	for step := 0; step < numSteps; step++ {
		// Calculate forces
		massGrid := DepositMassToGrid(particles, width, height)
		potentialGrid := SolvePoissonFFT(massGrid, width, height, gravitationalConstant)
		forceField := CalculateGradient(potentialGrid, width, height)

		// Apply leapfrog step
		LeapfrogStep(particles, forceField, dt, width, height)
	}

	// Center of mass should not move (no external forces)
	finalCOM := calculateCenterOfMass(particles)
	comDrift := finalCOM.Sub(initialCOM).Length()

	if comDrift > 0.01 {
		t.Errorf("Center of mass drifted: initial=%v, final=%v, drift=%v",
			initialCOM, finalCOM, comDrift)
	}
}

// Helper function to calculate center of mass
func calculateCenterOfMass(particles []*Particle) Vec3 {
	totalMass := float64(0)
	com := Vec3{0, 0, 0}

	for _, p := range particles {
		mass := float64(p.Mass)
		totalMass += mass
		com.X += p.Position.X * mass
		com.Y += p.Position.Y * mass
		com.Z += p.Position.Z * mass
	}

	if totalMass > 0 {
		com.X /= totalMass
		com.Y /= totalMass
		com.Z /= totalMass
	}

	return com
}
