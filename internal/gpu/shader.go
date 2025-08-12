package gpu

import (
	"errors"
	"fmt"
	"strings"
)

// ShaderManager manages compute shader compilation and caching
type ShaderManager struct {
	cache map[string]*ComputeShader
}

// NewShaderManager creates a new shader manager
func NewShaderManager() *ShaderManager {
	return &ShaderManager{
		cache: make(map[string]*ComputeShader),
	}
}

// CompileComputeShader compiles a compute shader from source
func (m *ShaderManager) CompileComputeShader(source string) (*ComputeShader, error) {
	// Without OpenGL context, we cannot actually compile
	// This is a placeholder that will be implemented when GPU support is added
	return nil, errors.New("OpenGL context not available")
}

// DeleteShader deletes a compiled shader
func (m *ShaderManager) DeleteShader(shader *ComputeShader) error {
	if shader == nil {
		return nil
	}

	// In real implementation, this would call OpenGL delete functions
	shader.ProgramID = 0
	return nil
}

// GetCacheSize returns the number of cached shaders
func (m *ShaderManager) GetCacheSize() int {
	return len(m.cache)
}

// CacheShader adds a shader to the cache
func (m *ShaderManager) CacheShader(key string, shader *ComputeShader) {
	m.cache[key] = shader
}

// GetCachedShader retrieves a shader from the cache
func (m *ShaderManager) GetCachedShader(key string) *ComputeShader {
	return m.cache[key]
}

// ClearCache removes all cached shaders
func (m *ShaderManager) ClearCache() {
	// In real implementation, we would delete all shaders first
	for _, shader := range m.cache {
		_ = m.DeleteShader(shader)
	}
	m.cache = make(map[string]*ComputeShader)
}

// GenerateFFTShader generates FFT compute shader source code
func (m *ShaderManager) GenerateFFTShader(width, height int, forward bool) string {
	direction := "1.0"
	if !forward {
		direction = "-1.0"
	}

	return fmt.Sprintf(`#version 430
layout (local_size_x = 8, local_size_y = 8, local_size_z = 1) in;

layout(std430, binding = 0) buffer InputBuffer {
    vec2 input_data[];
};

layout(std430, binding = 1) buffer OutputBuffer {
    vec2 output_data[];
};

uniform int u_width;
uniform int u_height;
uniform float u_direction = %s; // 1.0 for forward, -1.0 for inverse

void main() {
    uint x = gl_GlobalInvocationID.x;
    uint y = gl_GlobalInvocationID.y;
    
    if (x >= u_width || y >= u_height) return;
    
    // Simplified FFT computation placeholder
    uint idx = y * u_width + x;
    vec2 value = input_data[idx];
    
    // Apply direction for forward/inverse
    value *= u_direction;
    
    output_data[idx] = value;
}
`, direction)
}

// ValidateShaderSource validates compute shader source code
func (m *ShaderManager) ValidateShaderSource(source string) bool {
	if source == "" {
		return false
	}

	// Check for required components
	hasVersion := strings.Contains(source, "#version")
	hasMain := strings.Contains(source, "void main()")
	hasLayout := strings.Contains(source, "layout")

	return hasVersion && hasMain && hasLayout
}
