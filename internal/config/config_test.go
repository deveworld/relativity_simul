package config

import (
	"testing"
)

// TestDefaultConfig tests creating a default configuration
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test screen dimensions
	if cfg.ScreenWidth != 1920 {
		t.Errorf("Expected ScreenWidth 1920, got %d", cfg.ScreenWidth)
	}
	if cfg.ScreenHeight != 1080 {
		t.Errorf("Expected ScreenHeight 1080, got %d", cfg.ScreenHeight)
	}

	// Test simulation dimensions
	if cfg.SimulationWidth != 256 {
		t.Errorf("Expected SimulationWidth 256, got %d", cfg.SimulationWidth)
	}
	if cfg.SimulationDepth != 256 {
		t.Errorf("Expected SimulationDepth 256, got %d", cfg.SimulationDepth)
	}

	// Test physics parameters
	if cfg.NumParticles != 10 {
		t.Errorf("Expected NumParticles 10, got %d", cfg.NumParticles)
	}
	if cfg.GravitationalConstant != 1.0 {
		t.Errorf("Expected GravitationalConstant 1.0, got %f", cfg.GravitationalConstant)
	}

	// Test rendering parameters
	if cfg.GridVisScale != 0.1 {
		t.Errorf("Expected GridVisScale 0.1, got %f", cfg.GridVisScale)
	}
	if cfg.MoveSpeed != 0.3 {
		t.Errorf("Expected MoveSpeed 0.3, got %f", cfg.MoveSpeed)
	}
	if cfg.MouseSensitivity != 0.003 {
		t.Errorf("Expected MouseSensitivity 0.003, got %f", cfg.MouseSensitivity)
	}

	// Test initial camera settings
	if cfg.InitialYaw != 3.92699 {
		t.Errorf("Expected InitialYaw 3.92699, got %f", cfg.InitialYaw)
	}
	if cfg.InitialPitch != -0.628 {
		t.Errorf("Expected InitialPitch -0.628, got %f", cfg.InitialPitch)
	}

	// Test default flags
	if cfg.StartPaused != false {
		t.Errorf("Expected StartPaused false, got %v", cfg.StartPaused)
	}
	if cfg.UseGPU != true {
		t.Errorf("Expected UseGPU true, got %v", cfg.UseGPU)
	}
}

// TestCustomConfig tests creating a custom configuration
func TestCustomConfig(t *testing.T) {
	cfg := &Config{
		ScreenWidth:           1600,
		ScreenHeight:          900,
		SimulationWidth:       128,
		SimulationDepth:       128,
		NumParticles:          20,
		GravitationalConstant: 2.0,
		GridVisScale:          0.2,
		MoveSpeed:             0.5,
		MouseSensitivity:      0.005,
		InitialYaw:            0.0,
		InitialPitch:          0.0,
		StartPaused:           true,
		UseGPU:                false,
	}

	// Verify custom values
	if cfg.ScreenWidth != 1600 {
		t.Errorf("Expected ScreenWidth 1600, got %d", cfg.ScreenWidth)
	}
	if cfg.NumParticles != 20 {
		t.Errorf("Expected NumParticles 20, got %d", cfg.NumParticles)
	}
	if cfg.UseGPU != false {
		t.Errorf("Expected UseGPU false, got %v", cfg.UseGPU)
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name:      "valid config",
			config:    DefaultConfig(),
			wantError: false,
		},
		{
			name: "invalid screen width",
			config: &Config{
				ScreenWidth:     0,
				ScreenHeight:    1080,
				SimulationWidth: 256,
				SimulationDepth: 256,
				NumParticles:    10,
			},
			wantError: true,
		},
		{
			name: "invalid simulation dimensions",
			config: &Config{
				ScreenWidth:     1920,
				ScreenHeight:    1080,
				SimulationWidth: 0,
				SimulationDepth: 256,
				NumParticles:    10,
			},
			wantError: true,
		},
		{
			name: "invalid particle count",
			config: &Config{
				ScreenWidth:     1920,
				ScreenHeight:    1080,
				SimulationWidth: 256,
				SimulationDepth: 256,
				NumParticles:    -1,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
