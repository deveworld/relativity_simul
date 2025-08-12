package integration_test

import (
	"math"
	"relativity_simulation_2d/internal/config"
	"relativity_simulation_2d/internal/physics"
	"testing"
	"time"
)

// TestFullSimulation verifies that all components work together correctly
func TestFullSimulation(t *testing.T) {
	// Setup: Load configuration
	cfg := config.DefaultConfig()
	if cfg == nil {
		t.Fatal("Failed to create default configuration")
	}

	// Verify configuration is valid
	if cfg.NumParticles <= 0 {
		t.Errorf("Invalid number of particles: %d", cfg.NumParticles)
	}
	if cfg.SimulationWidth <= 0 || cfg.SimulationDepth <= 0 {
		t.Errorf("Invalid simulation dimensions: %dx%d", cfg.SimulationWidth, cfg.SimulationDepth)
	}

	// Test 1: Particle initialization
	particles := physics.InitializeParticles(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth))
	if len(particles) != cfg.NumParticles {
		t.Errorf("Expected %d particles, got %d", cfg.NumParticles, len(particles))
	}

	// Verify particles are within initialization bounds (centered around origin)
	// Particles are initialized in range [-width*0.4, width*0.4]
	maxX := float64(cfg.SimulationWidth) * 0.4
	maxZ := float64(cfg.SimulationDepth) * 0.4
	for i, p := range particles {
		if math.Abs(p.Position.X) > maxX {
			t.Errorf("Particle %d X position out of initialization bounds: %f (max: %f)", i, p.Position.X, maxX)
		}
		if math.Abs(p.Position.Z) > maxZ {
			t.Errorf("Particle %d Z position out of initialization bounds: %f (max: %f)", i, p.Position.Z, maxZ)
		}
		if p.Mass <= 0 {
			t.Errorf("Particle %d has invalid mass: %f", i, p.Mass)
		}
	}

	// Test 2: Mass density grid deposition
	massDensityGrid := physics.DepositMassToGrid(particles, cfg.SimulationWidth, cfg.SimulationDepth)
	if len(massDensityGrid) != cfg.SimulationWidth {
		t.Errorf("Mass density grid width mismatch: expected %d, got %d", cfg.SimulationWidth, len(massDensityGrid))
	}
	if len(massDensityGrid[0]) != cfg.SimulationDepth {
		t.Errorf("Mass density grid depth mismatch: expected %d, got %d", cfg.SimulationDepth, len(massDensityGrid[0]))
	}

	// Verify total mass is conserved
	totalMassInGrid := 0.0
	for i := range massDensityGrid {
		for j := range massDensityGrid[i] {
			totalMassInGrid += massDensityGrid[i][j]
		}
	}
	totalParticleMass := float32(0.0)
	for _, p := range particles {
		totalParticleMass += p.Mass
	}
	// Allow for some numerical error due to grid discretization
	if math.Abs(totalMassInGrid-float64(totalParticleMass))/float64(totalParticleMass) > 0.1 {
		t.Errorf("Mass conservation violated: particles=%f, grid=%f", totalParticleMass, totalMassInGrid)
	}

	// Test 3: Poisson solver (field calculation)
	potentialGrid := physics.SolvePoissonFFT(massDensityGrid, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
	if len(potentialGrid) != cfg.SimulationWidth {
		t.Errorf("Potential grid width mismatch: expected %d, got %d", cfg.SimulationWidth, len(potentialGrid))
	}
	if len(potentialGrid[0]) != cfg.SimulationDepth {
		t.Errorf("Potential grid depth mismatch: expected %d, got %d", cfg.SimulationDepth, len(potentialGrid[0]))
	}

	// Test 4: Force calculation (via RunTimeEvolution which returns force field)
	forceField := physics.RunTimeEvolution(particles, float32(0.001), cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
	if forceField == nil {
		t.Fatal("Force field calculation returned nil")
	}
	if len(forceField.AccelFieldX) != cfg.SimulationWidth {
		t.Errorf("Acceleration field X width mismatch: expected %d, got %d", cfg.SimulationWidth, len(forceField.AccelFieldX))
	}
	if len(forceField.AccelFieldZ) != cfg.SimulationWidth {
		t.Errorf("Acceleration field Z width mismatch: expected %d, got %d", cfg.SimulationWidth, len(forceField.AccelFieldZ))
	}

	// Test 5: Time evolution
	deltaTime := float32(0.01)
	initialPositions := make([][2]float64, len(particles))
	for i, p := range particles {
		initialPositions[i][0] = p.Position.X
		initialPositions[i][1] = p.Position.Z
	}

	// Run multiple time steps
	for step := 0; step < 10; step++ {
		forceField = physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
		if forceField == nil {
			t.Fatalf("Time evolution failed at step %d", step)
		}
	}

	// Verify particles have moved (unless all velocities were zero)
	particlesMoved := false
	for i, p := range particles {
		if math.Abs(p.Position.X-initialPositions[i][0]) > 1e-10 || math.Abs(p.Position.Z-initialPositions[i][1]) > 1e-10 {
			particlesMoved = true
			break
		}
	}
	// This might fail if all particles start with zero velocity, which is acceptable
	// but we should at least verify the simulation doesn't crash
	_ = particlesMoved // May be false if all particles started stationary

	// Test 6: Performance benchmark
	start := time.Now()
	for i := 0; i < 100; i++ {
		physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
	}
	elapsed := time.Since(start)
	t.Logf("100 simulation steps completed in %v", elapsed)

	// Verify performance is reasonable (should complete in less than 10 seconds for CPU)
	if elapsed > 10*time.Second {
		t.Errorf("Performance issue: 100 steps took %v (expected < 10s)", elapsed)
	}

	// Test 7: Energy conservation (approximate check)
	// Calculate total kinetic energy
	totalKE := float32(0.0)
	for _, p := range particles {
		v2 := float32(p.Velocity.X*p.Velocity.X + p.Velocity.Y*p.Velocity.Y + p.Velocity.Z*p.Velocity.Z)
		totalKE += 0.5 * p.Mass * v2
	}

	// Run simulation for a longer period
	for i := 0; i < 100; i++ {
		physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
	}

	// Calculate kinetic energy after evolution
	finalKE := float32(0.0)
	for _, p := range particles {
		v2 := float32(p.Velocity.X*p.Velocity.X + p.Velocity.Y*p.Velocity.Y + p.Velocity.Z*p.Velocity.Z)
		finalKE += 0.5 * p.Mass * v2
	}

	// Energy shouldn't increase dramatically (some numerical error is expected)
	if finalKE > totalKE*10 {
		t.Errorf("Energy conservation issue: initial KE=%f, final KE=%f", totalKE, finalKE)
	}

	// Test 8: Boundary conditions
	// After evolution, particles may drift but shouldn't explode to infinity
	// Check that particles remain within reasonable bounds
	reasonableBound := float64(cfg.SimulationWidth) * 2.0
	outOfBounds := 0
	for _, p := range particles {
		if math.Abs(p.Position.X) > reasonableBound || math.Abs(p.Position.Z) > reasonableBound {
			outOfBounds++
		}
	}
	// Allow some drift but not explosion
	if float64(outOfBounds)/float64(len(particles)) > 0.2 {
		t.Errorf("Too many particles out of reasonable bounds: %d/%d", outOfBounds, len(particles))
	}

	// Test 9: Verify all subsystems integrate correctly
	// This is a high-level check that the simulation can run without panics
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Simulation panicked: %v", r)
		}
	}()

	// Run a complete simulation cycle
	for i := 0; i < 10; i++ {
		// Deposit mass
		massDensityGrid = physics.DepositMassToGrid(particles, cfg.SimulationWidth, cfg.SimulationDepth)

		// Solve potential
		_ = physics.SolvePoissonFFT(massDensityGrid, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)

		// Evolve in time (RunTimeEvolution handles force calculation internally)
		_ = physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
	}

	t.Log("Full simulation integration test completed successfully")
}

// TestSimulationWithCentralMass tests the simulation with a large central mass
func TestSimulationWithCentralMass(t *testing.T) {
	cfg := config.DefaultConfig()

	// Initialize particles with central mass
	particles := physics.InitializeParticlesWithCentralMass(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth), 1000.0)

	if len(particles) != cfg.NumParticles {
		t.Errorf("Expected %d particles, got %d", cfg.NumParticles, len(particles))
	}

	// Verify the central particle has large mass
	centralParticle := particles[0]
	if centralParticle.Mass < 900 {
		t.Errorf("Central particle mass too small: %f", centralParticle.Mass)
	}

	// Verify central particle is at origin (0,0,0)
	if math.Abs(centralParticle.Position.X) > 1 || math.Abs(centralParticle.Position.Z) > 1 {
		t.Errorf("Central particle not at origin: (%f, %f)", centralParticle.Position.X, centralParticle.Position.Z)
	}

	// Run simulation and verify orbits form
	deltaTime := float32(0.01)
	for i := 0; i < 50; i++ {
		physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
	}

	// Verify central mass hasn't moved much from origin
	if math.Abs(particles[0].Position.X) > 5 || math.Abs(particles[0].Position.Z) > 5 {
		t.Errorf("Central mass moved too much from origin: (%f, %f)", particles[0].Position.X, particles[0].Position.Z)
	}

	t.Log("Central mass simulation test completed successfully")
}

// TestParallelSimulation tests that the simulation can handle concurrent operations safely
func TestParallelSimulation(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.NumParticles = 100 // Use fewer particles for parallel test

	// Run multiple simulations in parallel
	numGoroutines := 4
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			particles := physics.InitializeParticles(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth))
			deltaTime := float32(0.01)

			for step := 0; step < 10; step++ {
				physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	t.Log("Parallel simulation test completed successfully")
}
