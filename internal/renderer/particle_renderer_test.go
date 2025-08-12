package renderer

import (
	"relativity_simulation_2d/internal/physics"
	"testing"
)

// TestParticleRendererCreation tests creating a particle renderer
func TestParticleRendererCreation(t *testing.T) {
	renderer := NewParticleRenderer()

	if renderer == nil {
		t.Fatal("Failed to create particle renderer")
	}

	// Check default settings
	if renderer.GetParticleSize() == 0 {
		t.Error("Particle size should have a default value")
	}
}

// TestParticleRendererSetup tests setting up the renderer
func TestParticleRendererSetup(t *testing.T) {
	renderer := NewParticleRenderer()

	// Setup renderer (in real implementation, this would initialize shaders)
	err := renderer.Setup()
	if err != nil {
		// It's OK if setup fails without OpenGL context
		t.Logf("Setup failed (expected in test): %v", err)
	}
}

// TestAddParticles tests adding particles to render
func TestAddParticles(t *testing.T) {
	renderer := NewParticleRenderer()

	// Create test particles
	particles := []*physics.Particle{
		physics.NewParticle(1.0, 0, 0, 0, 0, 0, 0),
		physics.NewParticle(2.0, 10, 0, 0, 0, 0, 0),
		physics.NewParticle(3.0, 0, 0, 10, 0, 0, 0),
	}

	// Add particles to renderer
	renderer.SetParticles(particles)

	// Check particle count
	if renderer.GetParticleCount() != len(particles) {
		t.Errorf("Expected %d particles, got %d",
			len(particles), renderer.GetParticleCount())
	}
}

// TestRenderBatch tests batch rendering
func TestRenderBatch(t *testing.T) {
	renderer := NewParticleRenderer()

	// Create many particles for batch rendering
	numParticles := 1000
	particles := make([]*physics.Particle, numParticles)
	for i := 0; i < numParticles; i++ {
		particles[i] = physics.NewParticle(
			1.0,
			float64(i%10)*10,
			0,
			float64(i/10)*10,
			0, 0, 0,
		)
	}

	renderer.SetParticles(particles)

	// Test batch info
	batches := renderer.GetBatchInfo()
	if batches.TotalBatches == 0 {
		t.Error("Should have at least one batch")
	}

	if batches.ParticlesPerBatch == 0 {
		t.Error("Particles per batch should be non-zero")
	}

	// Total particles should match
	totalInBatches := batches.TotalBatches * batches.ParticlesPerBatch
	if totalInBatches < numParticles {
		t.Error("Batches don't cover all particles")
	}
}

// TestColorMapping tests particle color based on properties
func TestColorMapping(t *testing.T) {
	renderer := NewParticleRenderer()

	// Test color based on mass
	lightParticle := physics.NewParticle(1.0, 0, 0, 0, 0, 0, 0)
	heavyParticle := physics.NewParticle(100.0, 0, 0, 0, 0, 0, 0)

	lightColor := renderer.GetParticleColor(lightParticle)
	heavyColor := renderer.GetParticleColor(heavyParticle)

	// Colors should be different for different masses
	if lightColor.R == heavyColor.R &&
		lightColor.G == heavyColor.G &&
		lightColor.B == heavyColor.B {
		t.Error("Particles with different masses should have different colors")
	}
}

// TestParticleSize tests particle size calculation
func TestParticleSize(t *testing.T) {
	renderer := NewParticleRenderer()

	// Set base particle size
	renderer.SetParticleSize(2.0)
	if renderer.GetParticleSize() != 2.0 {
		t.Error("Failed to set particle size")
	}

	// Test size based on mass
	smallParticle := physics.NewParticle(1.0, 0, 0, 0, 0, 0, 0)
	largeParticle := physics.NewParticle(10.0, 0, 0, 0, 0, 0, 0)

	smallSize := renderer.GetScaledParticleSize(smallParticle)
	largeSize := renderer.GetScaledParticleSize(largeParticle)

	if largeSize <= smallSize {
		t.Error("Larger mass should result in larger particle size")
	}
}

// TestCulling tests frustum culling for particles
func TestCulling(t *testing.T) {
	renderer := NewParticleRenderer()

	// Create camera
	camera := NewCamera(
		physics.NewVec3(0, 0, 0),
		physics.NewVec3(0, 0, -1),
		physics.NewVec3(0, 1, 0),
	)
	camera.SetPerspective(60.0, 1.0, 1.0, 100.0)

	// Create particles - some visible, some not
	particles := []*physics.Particle{
		physics.NewParticle(1.0, 0, 0, -10, 0, 0, 0),   // Visible
		physics.NewParticle(1.0, 0, 0, 10, 0, 0, 0),    // Behind camera
		physics.NewParticle(1.0, 0, 0, -200, 0, 0, 0),  // Beyond far plane
		physics.NewParticle(1.0, 200, 0, -10, 0, 0, 0), // Outside frustum
	}

	renderer.SetParticles(particles)
	renderer.SetCamera(camera)

	// Enable culling
	renderer.EnableCulling(true)

	// Get visible particles
	visibleCount := renderer.GetVisibleParticleCount()

	// Should only render the first particle
	if visibleCount != 1 {
		t.Errorf("Expected 1 visible particle, got %d", visibleCount)
	}
}

// TestRenderMode tests different rendering modes
func TestRenderMode(t *testing.T) {
	renderer := NewParticleRenderer()

	// Test point sprites mode
	renderer.SetRenderMode(RenderModePoints)
	if renderer.GetRenderMode() != RenderModePoints {
		t.Error("Failed to set point sprite mode")
	}

	// Test sphere mode
	renderer.SetRenderMode(RenderModeSpheres)
	if renderer.GetRenderMode() != RenderModeSpheres {
		t.Error("Failed to set sphere mode")
	}

	// Test billboard mode
	renderer.SetRenderMode(RenderModeBillboards)
	if renderer.GetRenderMode() != RenderModeBillboards {
		t.Error("Failed to set billboard mode")
	}
}

// TestCleanup tests proper cleanup of renderer resources
func TestCleanup(t *testing.T) {
	renderer := NewParticleRenderer()

	// Add some particles
	particles := []*physics.Particle{
		physics.NewParticle(1.0, 0, 0, 0, 0, 0, 0),
	}
	renderer.SetParticles(particles)

	// Cleanup
	err := renderer.Cleanup()
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}

	// After cleanup, particle count should be 0
	if renderer.GetParticleCount() != 0 {
		t.Error("Particles not cleared after cleanup")
	}
}
