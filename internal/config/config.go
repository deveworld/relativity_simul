package config

import (
	"fmt"
)

// Config holds all configuration parameters for the simulation
type Config struct {
	// Display settings
	ScreenWidth  int
	ScreenHeight int

	// Simulation dimensions
	SimulationWidth int
	SimulationDepth int

	// Physics parameters
	NumParticles          int
	GravitationalConstant float64

	// Rendering parameters
	GridVisScale     float64
	MoveSpeed        float32
	MouseSensitivity float32

	// Camera initial settings
	InitialYaw   float32
	InitialPitch float32

	// Runtime flags
	StartPaused bool
	UseGPU      bool
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		// Display settings
		ScreenWidth:  1920,
		ScreenHeight: 1080,

		// Simulation dimensions
		SimulationWidth: 256,
		SimulationDepth: 256,

		// Physics parameters
		NumParticles:          10,
		GravitationalConstant: 1.0,

		// Rendering parameters
		GridVisScale:     0.1,
		MoveSpeed:        0.3,
		MouseSensitivity: 0.003,

		// Camera initial settings
		InitialYaw:   3.92699, // Start facing -Z direction
		InitialPitch: -0.628,  // Start looking slightly down

		// Runtime flags
		StartPaused: false,
		UseGPU:      true,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ScreenWidth <= 0 {
		return fmt.Errorf("invalid screen width: %d", c.ScreenWidth)
	}
	if c.ScreenHeight <= 0 {
		return fmt.Errorf("invalid screen height: %d", c.ScreenHeight)
	}
	if c.SimulationWidth <= 0 {
		return fmt.Errorf("invalid simulation width: %d", c.SimulationWidth)
	}
	if c.SimulationDepth <= 0 {
		return fmt.Errorf("invalid simulation depth: %d", c.SimulationDepth)
	}
	if c.NumParticles < 0 {
		return fmt.Errorf("invalid number of particles: %d", c.NumParticles)
	}
	return nil
}

// Clone creates a deep copy of the configuration
func (c *Config) Clone() *Config {
	clone := *c
	return &clone
}
