package gpu

import (
	"testing"
)

func TestGPUMemoryBufferCreation(t *testing.T) {
	// Test 1.5: Test for compute buffer creation

	// Create a GPU memory buffer with specific size
	bufferSize := 1024 * 1024 // 1MB
	buffer := &GPUMemoryBuffer{
		BufferID: 1,
		Size:     bufferSize,
	}

	// Verify buffer creation
	if buffer.BufferID == 0 {
		t.Error("Buffer ID should not be zero after creation")
	}

	if buffer.Size != bufferSize {
		t.Errorf("Buffer size mismatch: expected %d, got %d", bufferSize, buffer.Size)
	}
}

func TestComplexGPUBufferCreation(t *testing.T) {
	// Test creating a complex GPU buffer for FFT operations
	elementCount := 256
	buffer := &ComplexGPUBuffer{
		BufferID: 2,
		Size:     elementCount,
	}

	// Verify buffer creation
	if buffer.BufferID == 0 {
		t.Error("Complex buffer ID should not be zero after creation")
	}

	if buffer.Size != elementCount {
		t.Errorf("Complex buffer size mismatch: expected %d, got %d", elementCount, buffer.Size)
	}
}

func TestComputeShaderCreation(t *testing.T) {
	// Test compute shader creation
	shader := &ComputeShader{
		ProgramID: 3,
	}

	if shader.ProgramID == 0 {
		t.Error("Shader program ID should not be zero after creation")
	}
}

func TestGPUStructCreation(t *testing.T) {
	// Test GPU struct creation
	gpu := &GPU{
		Initialized: false,
		Headless:    false,
	}

	if gpu.Initialized {
		t.Error("GPU should not be initialized by default")
	}
}
