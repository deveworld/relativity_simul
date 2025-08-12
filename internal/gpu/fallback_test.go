package gpu

import (
	"testing"
)

// TestGPUFallbackManager tests the GPU fallback mechanism
func TestGPUFallbackManager(t *testing.T) {
	// Create a fallback manager
	manager := NewFallbackManager()

	// Test that manager is created successfully
	if manager == nil {
		t.Fatal("Failed to create fallback manager")
	}

	// Test default mode (should prefer GPU if available)
	if manager.GetMode() != ModeAuto {
		t.Errorf("Expected default mode to be Auto, got %v", manager.GetMode())
	}
}

// TestGPUDetection tests GPU availability detection
func TestGPUDetection(t *testing.T) {
	manager := NewFallbackManager()

	// Check GPU availability
	hasGPU := manager.IsGPUAvailable()

	// Log the result (won't fail test since GPU may not be available)
	t.Logf("GPU available: %v", hasGPU)

	// Test that we can query GPU info even if not available
	info := manager.GetGPUInfo()
	if info == nil {
		t.Error("GetGPUInfo should return info even if GPU not available")
	}
}

// TestFallbackToCPU tests falling back to CPU when GPU fails
func TestFallbackToCPU(t *testing.T) {
	manager := NewFallbackManager()

	// Force CPU mode
	manager.SetMode(ModeCPU)

	if manager.GetMode() != ModeCPU {
		t.Errorf("Failed to set CPU mode, got %v", manager.GetMode())
	}

	// Test that compute operations use CPU
	processor := manager.GetProcessor()
	if processor == nil {
		t.Fatal("Failed to get processor")
	}

	if processor.GetType() != ProcessorTypeCPU {
		t.Errorf("Expected CPU processor, got %v", processor.GetType())
	}
}

// TestFallbackOnGPUError tests automatic fallback on GPU error
func TestFallbackOnGPUError(t *testing.T) {
	manager := NewFallbackManager()

	// Start in GPU mode
	manager.SetMode(ModeGPU)

	// Simulate GPU error
	err := manager.SimulateGPUError()
	if err != nil {
		t.Errorf("Failed to simulate GPU error: %v", err)
	}

	// Check that it fell back to CPU
	if manager.GetMode() != ModeCPU {
		t.Error("Should have fallen back to CPU after GPU error")
	}

	// Verify error was logged
	if !manager.HasError() {
		t.Error("Error should be recorded")
	}

	lastError := manager.GetLastError()
	if lastError == nil {
		t.Error("Last error should not be nil")
	}
}

// TestProcessorSelection tests selecting the right processor
func TestProcessorSelection(t *testing.T) {
	manager := NewFallbackManager()

	testCases := []struct {
		mode     ComputeMode
		expected ProcessorType
	}{
		{ModeCPU, ProcessorTypeCPU},
		{ModeGPU, ProcessorTypeGPU},
		{ModeAuto, ProcessorTypeCPU}, // CPU when GPU not available
	}

	for _, tc := range testCases {
		t.Run(tc.mode.String(), func(t *testing.T) {
			manager.SetMode(tc.mode)
			processor := manager.GetProcessor()

			if processor == nil {
				t.Fatal("Processor should not be nil")
			}

			// When GPU is not available, GPU mode should fall back to CPU
			if tc.mode == ModeGPU && !manager.IsGPUAvailable() {
				if processor.GetType() != ProcessorTypeCPU {
					t.Errorf("Should fall back to CPU when GPU not available")
				}
			} else if processor.GetType() != tc.expected {
				t.Errorf("Expected processor type %v, got %v",
					tc.expected, processor.GetType())
			}
		})
	}
}

// TestPerformanceMonitoring tests performance monitoring for fallback decisions
func TestPerformanceMonitoring(t *testing.T) {
	manager := NewFallbackManager()

	// Record some performance metrics
	manager.RecordPerformance(ProcessorTypeCPU, 100.0)
	manager.RecordPerformance(ProcessorTypeGPU, 50.0)

	// Get performance stats
	stats := manager.GetPerformanceStats()
	if stats == nil {
		t.Fatal("Performance stats should not be nil")
	}

	// Check CPU stats
	cpuStats := stats.CPUStats
	if cpuStats.Count == 0 {
		t.Error("CPU stats not recorded")
	}

	// In Auto mode, should prefer faster processor
	manager.SetMode(ModeAuto)
	processor := manager.GetProcessor()

	// With mock data showing GPU is faster, it should prefer GPU if available
	// But since GPU is not available in tests, it should use CPU
	if processor.GetType() != ProcessorTypeCPU {
		t.Errorf("Expected CPU processor in test environment, got %v",
			processor.GetType())
	}
}

// TestFallbackRecovery tests recovering from fallback
func TestFallbackRecovery(t *testing.T) {
	manager := NewFallbackManager()

	// Simulate GPU error and fallback
	manager.SetMode(ModeGPU)
	_ = manager.SimulateGPUError()

	if manager.GetMode() != ModeCPU {
		t.Error("Should be in CPU mode after error")
	}

	// Try to recover
	err := manager.AttemptRecovery()
	if err != nil {
		// Recovery might fail if GPU not available, which is OK
		t.Logf("Recovery failed (expected in test): %v", err)
	}

	// Clear errors
	manager.ClearErrors()
	if manager.HasError() {
		t.Error("Errors should be cleared")
	}
}

// TestConcurrentFallback tests thread-safe fallback operations
func TestConcurrentFallback(t *testing.T) {
	manager := NewFallbackManager()

	// Run concurrent operations
	done := make(chan bool, 3)

	// Goroutine 1: Toggle modes
	go func() {
		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				manager.SetMode(ModeCPU)
			} else {
				manager.SetMode(ModeGPU)
			}
		}
		done <- true
	}()

	// Goroutine 2: Get processor
	go func() {
		for i := 0; i < 10; i++ {
			_ = manager.GetProcessor()
		}
		done <- true
	}()

	// Goroutine 3: Record performance
	go func() {
		for i := 0; i < 10; i++ {
			manager.RecordPerformance(ProcessorTypeCPU, float64(i))
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// If we get here without deadlock or panic, the test passes
	t.Log("Concurrent operations completed successfully")
}
