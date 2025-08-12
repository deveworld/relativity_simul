package simulation

import (
	"relativity_simulation_2d/internal/config"
	"relativity_simulation_2d/internal/gpu"
	"relativity_simulation_2d/internal/physics"
)

// Simulation holds the entire state of the GR simulation
type Simulation struct {
	Config           *config.Config
	Particles        []*physics.Particle
	PotentialGrid    [][]float64 // Stores the scalar potential Φ (proportional to h_00)
	MassDensityGrid  [][]float64 // Stores the mass density ρ
	AccelFieldX      [][]float64 // Stores the X component of the acceleration field
	AccelFieldZ      [][]float64 // Stores the Z component of the acceleration field
	gpu              *gpu.GPU    // Optional GPU context for acceleration (nil = CPU-only)
	gpuErrorOccurred bool        // Tracks if GPU error occurred
}

// NewSimulation creates and initializes a new simulation instance
func NewSimulation(cfg *config.Config) *Simulation {
	sim := &Simulation{
		Config:          cfg,
		Particles:       make([]*physics.Particle, cfg.NumParticles),
		PotentialGrid:   make([][]float64, cfg.SimulationWidth),
		MassDensityGrid: make([][]float64, cfg.SimulationWidth),
		AccelFieldX:     make([][]float64, cfg.SimulationWidth),
		AccelFieldZ:     make([][]float64, cfg.SimulationWidth),
	}

	for i := range sim.PotentialGrid {
		sim.PotentialGrid[i] = make([]float64, cfg.SimulationDepth)
		sim.MassDensityGrid[i] = make([]float64, cfg.SimulationDepth)
		sim.AccelFieldX[i] = make([]float64, cfg.SimulationDepth)
		sim.AccelFieldZ[i] = make([]float64, cfg.SimulationDepth)
	}

	// Initialize particles using extracted function
	sim.Particles = physics.InitializeParticles(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth))

	// Optionally add a large central mass (uncomment to enable)
	// sim.Particles = physics.InitializeParticlesWithCentralMass(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth), 1000)

	return sim
}

// SetGPU sets the GPU context for acceleration
func (s *Simulation) SetGPU(gpuCtx *gpu.GPU) {
	s.gpu = gpuCtx
}

// CleanupGPU releases GPU resources if allocated
func (s *Simulation) CleanupGPU() {
	if s.gpu != nil {
		// GPU cleanup would be handled by the GPU package
		s.gpu = nil
	}
}

// HasGPUErrorOccurred returns true if a GPU error was encountered
func (s *Simulation) HasGPUErrorOccurred() bool {
	return s.gpuErrorOccurred
}

// Update runs one full step of the simulation with frame-rate independent timing
func (s *Simulation) Update(deltaTime float32) {
	// Use the extracted physics engine for time evolution
	forceField := physics.RunTimeEvolution(s.Particles, deltaTime, s.Config.SimulationWidth, s.Config.SimulationDepth, s.Config.GravitationalConstant)

	// Update our internal acceleration fields for visualization
	s.AccelFieldX = forceField.AccelFieldX
	s.AccelFieldZ = forceField.AccelFieldZ

	// Update mass density grid for visualization
	s.MassDensityGrid = physics.DepositMassToGrid(s.Particles, s.Config.SimulationWidth, s.Config.SimulationDepth)

	// Update potential grid for visualization
	s.PotentialGrid = physics.SolvePoissonFFT(s.MassDensityGrid, s.Config.SimulationWidth, s.Config.SimulationDepth, s.Config.GravitationalConstant)
}

// GetParticles returns the current particles
func (s *Simulation) GetParticles() []*physics.Particle {
	return s.Particles
}

// GetPotentialGrid returns the potential grid for visualization
func (s *Simulation) GetPotentialGrid() [][]float64 {
	return s.PotentialGrid
}

// GetConfig returns the simulation configuration
func (s *Simulation) GetConfig() *config.Config {
	return s.Config
}
