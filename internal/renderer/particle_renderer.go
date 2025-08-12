package renderer

import (
	"errors"
	"math"
	"relativity_simulation_2d/internal/physics"
)

// RenderMode represents the particle rendering mode
type RenderMode int

const (
	// RenderModePoints renders particles as points
	RenderModePoints RenderMode = iota
	// RenderModeSpheres renders particles as spheres
	RenderModeSpheres
	// RenderModeBillboards renders particles as billboards
	RenderModeBillboards
)

// Color represents an RGBA color
type Color struct {
	R, G, B, A float32
}

// BatchInfo contains batch rendering information
type BatchInfo struct {
	TotalBatches      int
	ParticlesPerBatch int
}

// ParticleRenderer handles rendering of particles
type ParticleRenderer struct {
	particles      []*physics.Particle
	camera         *Camera
	particleSize   float32
	renderMode     RenderMode
	cullingEnabled bool

	// Render state
	visibleCount int
	maxBatchSize int
}

// NewParticleRenderer creates a new particle renderer
func NewParticleRenderer() *ParticleRenderer {
	return &ParticleRenderer{
		particles:    make([]*physics.Particle, 0),
		particleSize: 1.0,
		renderMode:   RenderModePoints,
		maxBatchSize: 1000,
	}
}

// Setup initializes the renderer
func (r *ParticleRenderer) Setup() error {
	// In a real implementation, this would initialize shaders
	// For now, return an error since we don't have OpenGL context
	return errors.New("OpenGL context not available")
}

// SetParticles sets the particles to render
func (r *ParticleRenderer) SetParticles(particles []*physics.Particle) {
	r.particles = particles
	r.updateVisibleCount()
}

// GetParticleCount returns the number of particles
func (r *ParticleRenderer) GetParticleCount() int {
	return len(r.particles)
}

// GetParticleSize returns the base particle size
func (r *ParticleRenderer) GetParticleSize() float32 {
	return r.particleSize
}

// SetParticleSize sets the base particle size
func (r *ParticleRenderer) SetParticleSize(size float32) {
	r.particleSize = size
}

// GetBatchInfo returns batch rendering information
func (r *ParticleRenderer) GetBatchInfo() BatchInfo {
	if len(r.particles) == 0 {
		return BatchInfo{TotalBatches: 0, ParticlesPerBatch: 0}
	}

	totalBatches := (len(r.particles) + r.maxBatchSize - 1) / r.maxBatchSize
	return BatchInfo{
		TotalBatches:      totalBatches,
		ParticlesPerBatch: r.maxBatchSize,
	}
}

// GetParticleColor returns the color for a particle based on its properties
func (r *ParticleRenderer) GetParticleColor(particle *physics.Particle) Color {
	// Map mass to color - lighter particles are bluish, heavier are reddish
	massNorm := math.Min(float64(particle.Mass)/100.0, 1.0)

	return Color{
		R: float32(massNorm),
		G: float32(0.5),
		B: float32(1.0 - massNorm),
		A: 1.0,
	}
}

// GetScaledParticleSize returns the scaled size for a particle based on its mass
func (r *ParticleRenderer) GetScaledParticleSize(particle *physics.Particle) float32 {
	// Scale based on cube root of mass (volume scaling)
	massScale := float32(math.Pow(float64(particle.Mass), 1.0/3.0))
	return r.particleSize * massScale
}

// SetCamera sets the camera for culling
func (r *ParticleRenderer) SetCamera(camera *Camera) {
	r.camera = camera
	r.updateVisibleCount()
}

// EnableCulling enables or disables frustum culling
func (r *ParticleRenderer) EnableCulling(enable bool) {
	r.cullingEnabled = enable
	r.updateVisibleCount()
}

// GetVisibleParticleCount returns the number of visible particles
func (r *ParticleRenderer) GetVisibleParticleCount() int {
	return r.visibleCount
}

// updateVisibleCount updates the count of visible particles
func (r *ParticleRenderer) updateVisibleCount() {
	if !r.cullingEnabled || r.camera == nil {
		r.visibleCount = len(r.particles)
		return
	}

	count := 0
	for _, particle := range r.particles {
		if r.camera.IsPointInFrustum(particle.Position) {
			count++
		}
	}
	r.visibleCount = count
}

// SetRenderMode sets the rendering mode
func (r *ParticleRenderer) SetRenderMode(mode RenderMode) {
	r.renderMode = mode
}

// GetRenderMode returns the current rendering mode
func (r *ParticleRenderer) GetRenderMode() RenderMode {
	return r.renderMode
}

// Render renders all particles
func (r *ParticleRenderer) Render() error {
	if r.camera == nil {
		return errors.New("camera not set")
	}

	// In a real implementation, this would:
	// 1. Bind shaders
	// 2. Set uniforms (view, projection matrices)
	// 3. Upload particle data to GPU
	// 4. Draw particles based on render mode

	// For now, this is a no-op
	return nil
}

// RenderBatch renders a batch of particles
func (r *ParticleRenderer) RenderBatch(batchIndex int) error {
	batchInfo := r.GetBatchInfo()
	if batchIndex >= batchInfo.TotalBatches {
		return errors.New("batch index out of range")
	}

	// Calculate batch range
	// startIdx := batchIndex * r.maxBatchSize
	// endIdx := startIdx + r.maxBatchSize
	// if endIdx > len(r.particles) {
	// 	endIdx = len(r.particles)
	// }

	// In a real implementation, render particles[startIdx:endIdx]
	return nil
}

// Cleanup releases renderer resources
func (r *ParticleRenderer) Cleanup() error {
	// Clear particles
	r.particles = r.particles[:0]
	r.visibleCount = 0

	// In a real implementation, this would release GPU resources
	return nil
}

// GetVisibleParticles returns a list of visible particles
func (r *ParticleRenderer) GetVisibleParticles() []*physics.Particle {
	if !r.cullingEnabled || r.camera == nil {
		return r.particles
	}

	visible := make([]*physics.Particle, 0, r.visibleCount)
	for _, particle := range r.particles {
		if r.camera.IsPointInFrustum(particle.Position) {
			visible = append(visible, particle)
		}
	}

	return visible
}

// SetMaxBatchSize sets the maximum batch size
func (r *ParticleRenderer) SetMaxBatchSize(size int) {
	if size > 0 {
		r.maxBatchSize = size
	}
}
