package physics

import (
	"math"
	"testing"
)

func TestInitializeParticles(t *testing.T) {
	// Test 2.1: Test particle initialization with random distribution

	// Test parameters
	numParticles := 100
	simulationWidth := 200.0
	simulationHeight := 200.0

	// Initialize particles
	particles := InitializeParticles(numParticles, simulationWidth, simulationHeight)

	// Verify correct number of particles
	if len(particles) != numParticles {
		t.Errorf("Expected %d particles, got %d", numParticles, len(particles))
	}

	// Verify all particles are within bounds
	for i, p := range particles {
		if p == nil {
			t.Errorf("Particle %d is nil", i)
			continue
		}

		// Check position bounds (80% of simulation width/height as per original)
		maxX := simulationWidth * 0.8 / 2.0
		maxZ := simulationHeight * 0.8 / 2.0

		if math.Abs(p.Position.X) > maxX {
			t.Errorf("Particle %d X position out of bounds: %f > %f", i, p.Position.X, maxX)
		}

		if p.Position.Y != 0 {
			t.Errorf("Particle %d Y position should be 0, got %f", i, p.Position.Y)
		}

		if math.Abs(p.Position.Z) > maxZ {
			t.Errorf("Particle %d Z position out of bounds: %f > %f", i, p.Position.Z, maxZ)
		}

		// Check mass is in expected range (20.0 to 50.0)
		if p.Mass < 20.0 || p.Mass > 50.0 {
			t.Errorf("Particle %d mass out of range: %f", i, p.Mass)
		}

		// Check radius is calculated correctly
		expectedRadius := float32(math.Pow(float64(p.Mass/20.0), 1.0/3.0)) * 0.5
		if math.Abs(float64(p.Radius-expectedRadius)) > 0.001 {
			t.Errorf("Particle %d radius incorrect: got %f, expected %f", i, p.Radius, expectedRadius)
		}

		// Check velocity is zero (initial velocities)
		if p.Velocity.X != 0 || p.Velocity.Y != 0 || p.Velocity.Z != 0 {
			t.Errorf("Particle %d velocity should be zero, got: (%f, %f, %f)", i, p.Velocity.X, p.Velocity.Y, p.Velocity.Z)
		}
	}
}

func TestInitializeParticlesWithCentralMass(t *testing.T) {
	// Test particle initialization with a large central mass

	numParticles := 50
	simulationWidth := 200.0
	simulationHeight := 200.0
	centralMass := 1000.0

	particles := InitializeParticlesWithCentralMass(numParticles, simulationWidth, simulationHeight, centralMass)

	// Verify correct number of particles
	if len(particles) != numParticles {
		t.Errorf("Expected %d particles, got %d", numParticles, len(particles))
	}

	// Check first particle is the central mass
	if particles[0].Mass != float32(centralMass) {
		t.Errorf("Central mass incorrect: got %f, expected %f", particles[0].Mass, centralMass)
	}

	// Central mass should be at origin
	if particles[0].Position.X != 0 || particles[0].Position.Y != 0 || particles[0].Position.Z != 0 {
		t.Errorf("Central mass not at origin: (%f, %f, %f)",
			particles[0].Position.X, particles[0].Position.Y, particles[0].Position.Z)
	}

	// Central mass should have zero velocity
	if particles[0].Velocity.X != 0 || particles[0].Velocity.Y != 0 || particles[0].Velocity.Z != 0 {
		t.Errorf("Central mass has non-zero velocity: (%f, %f, %f)",
			particles[0].Velocity.X, particles[0].Velocity.Y, particles[0].Velocity.Z)
	}
}

func TestParticleDistribution(t *testing.T) {
	// Test that particles are reasonably distributed (not all clumped together)

	numParticles := 1000
	simulationWidth := 200.0
	simulationHeight := 200.0

	particles := InitializeParticles(numParticles, simulationWidth, simulationHeight)

	// Divide space into quadrants and count particles
	quadrants := make([]int, 4) // [+X+Z, -X+Z, -X-Z, +X-Z]

	for _, p := range particles {
		quadIndex := 0
		if p.Position.X < 0 {
			quadIndex += 1
		}
		if p.Position.Z < 0 {
			quadIndex += 2
		}
		quadrants[quadIndex]++
	}

	// Each quadrant should have roughly 25% of particles (with some tolerance)
	expectedPerQuadrant := numParticles / 4
	tolerance := float64(expectedPerQuadrant) * 0.3 // 30% tolerance

	for i, count := range quadrants {
		deviation := math.Abs(float64(count - expectedPerQuadrant))
		if deviation > tolerance {
			t.Errorf("Quadrant %d has uneven distribution: %d particles (expected ~%d)",
				i, count, expectedPerQuadrant)
		}
	}
}
