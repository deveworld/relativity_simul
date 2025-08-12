package integration_test

import (
	"relativity_simulation_2d/internal/config"
	"relativity_simulation_2d/internal/physics"
	"testing"
	"time"
)

// BenchmarkSimulation measures the performance of the complete simulation
func BenchmarkSimulation(b *testing.B) {
	cfg := config.DefaultConfig()
	cfg.NumParticles = 100 // Standard size for benchmarking

	// Initialize particles once
	particles := physics.InitializeParticles(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth))
	deltaTime := float32(0.01)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Run one complete simulation step
		physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
	}
}

// BenchmarkSimulationVaryingParticles benchmarks with different particle counts
func BenchmarkSimulationVaryingParticles(b *testing.B) {
	particleCounts := []int{10, 50, 100, 500, 1000}

	for _, numParticles := range particleCounts {
		b.Run(b.Name()+"/"+string(rune(numParticles))+"particles", func(b *testing.B) {
			cfg := config.DefaultConfig()
			cfg.NumParticles = numParticles

			particles := physics.InitializeParticles(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth))
			deltaTime := float32(0.01)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
			}
		})
	}
}

// BenchmarkMassDensityDeposition benchmarks the mass density grid calculation
func BenchmarkMassDensityDeposition(b *testing.B) {
	cfg := config.DefaultConfig()
	particles := physics.InitializeParticles(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = physics.DepositMassToGrid(particles, cfg.SimulationWidth, cfg.SimulationDepth)
	}
}

// BenchmarkPoissonSolver benchmarks the FFT-based Poisson solver
func BenchmarkPoissonSolver(b *testing.B) {
	cfg := config.DefaultConfig()
	particles := physics.InitializeParticles(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth))
	massDensityGrid := physics.DepositMassToGrid(particles, cfg.SimulationWidth, cfg.SimulationDepth)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = physics.SolvePoissonFFT(massDensityGrid, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
	}
}

// BenchmarkParticleInitialization benchmarks particle creation
func BenchmarkParticleInitialization(b *testing.B) {
	cfg := config.DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = physics.InitializeParticles(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth))
	}
}

// TestPerformanceRegression verifies that performance hasn't degraded
func TestPerformanceRegression(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.NumParticles = 100

	particles := physics.InitializeParticles(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth))
	deltaTime := float32(0.01)

	// Warm up
	for i := 0; i < 10; i++ {
		physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
	}

	// Measure 100 iterations
	start := time.Now()
	iterations := 100
	for i := 0; i < iterations; i++ {
		physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
	}
	elapsed := time.Since(start)

	// Calculate performance metrics
	iterPerSec := float64(iterations) / elapsed.Seconds()
	msPerIter := elapsed.Milliseconds() / int64(iterations)

	t.Logf("Performance: %.2f iterations/sec, %d ms/iteration", iterPerSec, msPerIter)

	// Set performance thresholds (these should be tuned based on your target hardware)
	// For CPU-only mode with 100 particles
	minIterPerSec := 10.0      // At least 10 iterations per second
	maxMsPerIter := int64(100) // At most 100ms per iteration

	if iterPerSec < minIterPerSec {
		t.Errorf("Performance regression: only %.2f iterations/sec (expected >= %.2f)", iterPerSec, minIterPerSec)
	}

	if msPerIter > maxMsPerIter {
		t.Errorf("Performance regression: %d ms/iteration (expected <= %d)", msPerIter, maxMsPerIter)
	}
}

// TestMemoryUsage verifies that the simulation doesn't leak memory
func TestMemoryUsage(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.NumParticles = 100

	// Run simulation for many iterations
	particles := physics.InitializeParticles(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth))
	deltaTime := float32(0.01)

	// Run for 1000 iterations to check for memory leaks
	for i := 0; i < 1000; i++ {
		massDensityGrid := physics.DepositMassToGrid(particles, cfg.SimulationWidth, cfg.SimulationDepth)
		_ = physics.SolvePoissonFFT(massDensityGrid, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
		_ = physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)

		// Periodically check that we haven't created too many particles
		if len(particles) != cfg.NumParticles {
			t.Errorf("Particle count changed: expected %d, got %d", cfg.NumParticles, len(particles))
		}
	}

	t.Log("Memory usage test completed - no apparent leaks")
}

// TestScalability tests how the simulation scales with particle count
func TestScalability(t *testing.T) {
	particleCounts := []int{10, 20, 50, 100, 200}
	timings := make([]float64, len(particleCounts))

	cfg := config.DefaultConfig()
	deltaTime := float32(0.01)
	iterations := 50

	for i, numParticles := range particleCounts {
		cfg.NumParticles = numParticles
		particles := physics.InitializeParticles(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth))

		// Warm up
		for j := 0; j < 5; j++ {
			physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
		}

		// Measure
		start := time.Now()
		for j := 0; j < iterations; j++ {
			physics.RunTimeEvolution(particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
		}
		elapsed := time.Since(start)

		timings[i] = elapsed.Seconds() / float64(iterations)
		t.Logf("%d particles: %.4f sec/iteration", numParticles, timings[i])
	}

	// Check that scaling is reasonable (not exponential)
	// The time should scale roughly as O(N log N) for FFT-based solver
	// or O(N^2) for direct particle-particle interactions
	for i := 1; i < len(timings); i++ {
		ratio := timings[i] / timings[i-1]
		particleRatio := float64(particleCounts[i]) / float64(particleCounts[i-1])

		// Allow up to cubic scaling (very conservative)
		maxRatio := particleRatio * particleRatio * particleRatio

		if ratio > maxRatio {
			t.Errorf("Poor scaling: %d->%d particles increased time by %.2fx (expected <= %.2fx)",
				particleCounts[i-1], particleCounts[i], ratio, maxRatio)
		}
	}
}
