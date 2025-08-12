package gpu

import (
	"strings"
	"testing"
)

// TestShaderManager tests the shader management functionality
func TestShaderManager(t *testing.T) {
	// Create a shader manager
	manager := NewShaderManager()

	// Test that manager is created successfully
	if manager == nil {
		t.Fatal("Failed to create shader manager")
	}
}

// TestShaderCompilation tests shader compilation
func TestShaderCompilation(t *testing.T) {
	manager := NewShaderManager()

	// Test compiling a simple compute shader
	shaderSource := `
		#version 430
		layout (local_size_x = 1, local_size_y = 1, local_size_z = 1) in;
		void main() {
			// Simple compute shader that does nothing
		}
	`

	shader, err := manager.CompileComputeShader(shaderSource)

	// Note: This will fail without OpenGL context
	// In real GPU environment, this would compile
	if err == nil {
		defer func() { _ = manager.DeleteShader(shader) }()

		if shader.ProgramID == 0 {
			t.Error("Compiled shader has invalid ID")
		}
	} else {
		// Expected to fail without OpenGL context
		if !strings.Contains(err.Error(), "OpenGL context") &&
			!strings.Contains(err.Error(), "not available") {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

// TestShaderCache tests shader caching mechanism
func TestShaderCache(t *testing.T) {
	manager := NewShaderManager()

	// Test that cache is initialized
	if manager.GetCacheSize() != 0 {
		t.Errorf("Expected empty cache, got %d entries", manager.GetCacheSize())
	}

	// Test adding to cache
	testShader := &ComputeShader{ProgramID: 123}
	manager.CacheShader("test_shader", testShader)

	if manager.GetCacheSize() != 1 {
		t.Errorf("Expected 1 cached shader, got %d", manager.GetCacheSize())
	}

	// Test retrieving from cache
	cached := manager.GetCachedShader("test_shader")
	if cached == nil {
		t.Error("Failed to retrieve cached shader")
	} else if cached.ProgramID != testShader.ProgramID {
		t.Errorf("Cached shader mismatch: expected ID %d, got %d",
			testShader.ProgramID, cached.ProgramID)
	}

	// Test cache miss
	notFound := manager.GetCachedShader("non_existent")
	if notFound != nil {
		t.Error("Expected nil for non-existent shader")
	}
}

// TestShaderSourceGeneration tests generating shader source code
func TestShaderSourceGeneration(t *testing.T) {
	manager := NewShaderManager()

	// Test FFT shader generation
	fftSource := manager.GenerateFFTShader(64, 64, true)
	if fftSource == "" {
		t.Error("Failed to generate FFT shader source")
	}

	// Check that it contains expected elements
	if !strings.Contains(fftSource, "#version 430") {
		t.Error("FFT shader missing version directive")
	}
	if !strings.Contains(fftSource, "layout") {
		t.Error("FFT shader missing layout specification")
	}
	if !strings.Contains(fftSource, "void main()") {
		t.Error("FFT shader missing main function")
	}

	// Test inverse FFT shader generation
	ifftSource := manager.GenerateFFTShader(64, 64, false)
	if ifftSource == "" {
		t.Error("Failed to generate inverse FFT shader source")
	}

	// Forward and inverse should be different
	if fftSource == ifftSource {
		t.Error("Forward and inverse FFT shaders should be different")
	}
}

// TestShaderCleanup tests proper cleanup of shader resources
func TestShaderCleanup(t *testing.T) {
	manager := NewShaderManager()

	// Create mock shaders
	shader1 := &ComputeShader{ProgramID: 1}
	shader2 := &ComputeShader{ProgramID: 2}

	// Add to cache
	manager.CacheShader("shader1", shader1)
	manager.CacheShader("shader2", shader2)

	// Clear all shaders
	manager.ClearCache()

	if manager.GetCacheSize() != 0 {
		t.Errorf("Cache not cleared: still has %d entries", manager.GetCacheSize())
	}
}

// TestShaderValidation tests shader source validation
func TestShaderValidation(t *testing.T) {
	manager := NewShaderManager()

	testCases := []struct {
		name     string
		source   string
		expected bool
	}{
		{
			name:     "Valid compute shader",
			source:   "#version 430\nlayout (local_size_x = 1) in;\nvoid main() {}",
			expected: true,
		},
		{
			name:     "Missing version",
			source:   "layout (local_size_x = 1) in;\nvoid main() {}",
			expected: false,
		},
		{
			name:     "Missing main function",
			source:   "#version 430\nlayout (local_size_x = 1) in;",
			expected: false,
		},
		{
			name:     "Empty source",
			source:   "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid := manager.ValidateShaderSource(tc.source)
			if valid != tc.expected {
				t.Errorf("Validation failed for %s: expected %v, got %v",
					tc.name, tc.expected, valid)
			}
		})
	}
}
