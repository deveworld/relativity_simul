package integration_test

import (
	"relativity_simulation_2d/internal/config"
	"relativity_simulation_2d/internal/gpu"
	"relativity_simulation_2d/internal/physics"
	"testing"
	"time"
)

// TestGPUAcceleration verifies GPU-specific functionality
func TestGPUAcceleration(t *testing.T) {
	// Test 1: Fallback manager initialization
	t.Run("Fallback manager", func(t *testing.T) {
		fallbackMgr := gpu.NewFallbackManager()
		if fallbackMgr == nil {
			t.Fatal("Failed to create fallback manager")
		}

		// Check current compute mode
		mode := fallbackMgr.GetMode()
		t.Logf("Current compute mode: %v", mode)

		// Check if GPU is available
		if fallbackMgr.IsGPUAvailable() {
			t.Log("GPU is available")

			// Get GPU info
			gpuInfo := fallbackMgr.GetGPUInfo()
			if gpuInfo != nil {
				t.Logf("GPU Info: Name=%s, Memory=%d",
					gpuInfo.Name, gpuInfo.Memory)
			}
		} else {
			t.Log("GPU is not available, will use CPU")
		}

		// Get current processor
		processor := fallbackMgr.GetProcessor()
		if processor != nil {
			t.Logf("Using processor type: %v", processor.GetType())
		}
	})

	// Test 2: Buffer manager
	t.Run("Buffer manager", func(t *testing.T) {
		bufferMgr := gpu.NewBufferManager()
		if bufferMgr == nil {
			t.Fatal("Failed to create buffer manager")
		}

		// Create float buffer
		size := 1024
		buffer, err := bufferMgr.CreateFloatBuffer(size)
		if err != nil {
			t.Errorf("Failed to create float buffer: %v", err)
			return
		}
		defer func() {
			_ = bufferMgr.FreeBuffer(buffer)
		}()

		// Test data transfer
		testData := make([]float32, size)
		for i := range testData {
			testData[i] = float32(i)
		}

		err = bufferMgr.UploadFloatData(buffer, testData)
		if err != nil {
			t.Errorf("Failed to upload data: %v", err)
			return
		}

		// Download data back
		downloadedData, err := bufferMgr.DownloadFloatData(buffer, size)
		if err != nil {
			t.Errorf("Failed to download data: %v", err)
			return
		}

		// Verify data integrity
		for i := range testData {
			if downloadedData[i] != testData[i] {
				t.Errorf("Data mismatch at index %d: expected %f, got %f",
					i, testData[i], downloadedData[i])
				break
			}
		}

		t.Log("Buffer manager test passed")
	})

	// Test 3: Complex buffer management
	t.Run("Complex buffer", func(t *testing.T) {
		bufferMgr := gpu.NewBufferManager()
		if bufferMgr == nil {
			t.Fatal("Failed to create buffer manager")
		}

		// Create complex buffer for FFT
		size := 256
		buffer, err := bufferMgr.CreateComplexBuffer(size)
		if err != nil {
			t.Errorf("Failed to create complex buffer: %v", err)
			return
		}
		defer func() {
			_ = bufferMgr.FreeComplexBuffer(buffer)
		}()

		// Create test complex data
		testData := make([]complex128, size)
		for i := range testData {
			testData[i] = complex(float64(i), float64(-i))
		}

		err = bufferMgr.UploadComplexData(buffer, testData)
		if err != nil {
			t.Errorf("Failed to upload complex data: %v", err)
			return
		}

		// Download data back
		downloadedData, err := bufferMgr.DownloadComplexData(buffer, size)
		if err != nil {
			t.Errorf("Failed to download complex data: %v", err)
			return
		}

		// Verify data integrity
		for i := range testData {
			if real(downloadedData[i]) != real(testData[i]) ||
				imag(downloadedData[i]) != imag(testData[i]) {
				t.Errorf("Complex data mismatch at index %d", i)
				break
			}
		}

		t.Log("Complex buffer test passed")
	})

	// Test 4: Shader manager
	t.Run("Shader manager", func(t *testing.T) {
		shaderMgr := gpu.NewShaderManager()
		if shaderMgr == nil {
			t.Fatal("Failed to create shader manager")
		}

		// Generate FFT shader
		shaderSource := shaderMgr.GenerateFFTShader(256, 256, true)
		if shaderSource == "" {
			t.Error("Failed to generate FFT shader source")
			return
		}

		// Validate shader source
		if !shaderMgr.ValidateShaderSource(shaderSource) {
			t.Error("Generated shader source is invalid")
		}

		// Compile shader
		shader, err := shaderMgr.CompileComputeShader(shaderSource)
		if err != nil {
			// This may fail if OpenGL context is not available
			t.Logf("Shader compilation skipped (no OpenGL context): %v", err)
		} else if shader != nil {
			// Cache the shader
			shaderMgr.CacheShader("fft_test", shader)

			// Verify cache
			cachedShader := shaderMgr.GetCachedShader("fft_test")
			if cachedShader != shader {
				t.Error("Shader cache failed")
			}

			// Clean up
			_ = shaderMgr.DeleteShader(shader)
		}

		t.Log("Shader manager test passed")
	})

	// Test 5: Fallback mechanism
	t.Run("Fallback mechanism", func(t *testing.T) {
		fallbackMgr := gpu.NewFallbackManager()
		if fallbackMgr == nil {
			t.Fatal("Failed to create fallback manager")
		}

		// Simulate GPU error
		err := fallbackMgr.SimulateGPUError()
		if err == nil {
			t.Error("Expected error from SimulateGPUError")
		}

		// Check that error is recorded
		if !fallbackMgr.HasError() {
			t.Error("Expected HasError to return true after GPU error")
		}

		// Get last error
		lastErr := fallbackMgr.GetLastError()
		if lastErr == nil {
			t.Error("Expected GetLastError to return an error")
		}

		// Current mode should be CPU after error
		mode := fallbackMgr.GetMode()
		if mode != gpu.ModeCPU {
			t.Errorf("Expected CPU mode after GPU error, got %v", mode)
		}

		// Attempt recovery
		recoveryErr := fallbackMgr.AttemptRecovery()
		if recoveryErr != nil {
			t.Logf("Recovery failed (expected in test environment): %v", recoveryErr)
		}

		// Clear errors
		fallbackMgr.ClearErrors()
		if fallbackMgr.HasError() {
			t.Error("Expected errors to be cleared")
		}

		t.Log("Fallback mechanism test passed")
	})

	// Test 6: Performance monitoring
	t.Run("Performance monitoring", func(t *testing.T) {
		fallbackMgr := gpu.NewFallbackManager()
		if fallbackMgr == nil {
			t.Fatal("Failed to create fallback manager")
		}

		// Record some performance data
		fallbackMgr.RecordPerformance(gpu.ProcessorTypeCPU, 50.0)
		fallbackMgr.RecordPerformance(gpu.ProcessorTypeCPU, 45.0)
		fallbackMgr.RecordPerformance(gpu.ProcessorTypeCPU, 55.0)

		fallbackMgr.RecordPerformance(gpu.ProcessorTypeGPU, 20.0)
		fallbackMgr.RecordPerformance(gpu.ProcessorTypeGPU, 25.0)
		fallbackMgr.RecordPerformance(gpu.ProcessorTypeGPU, 22.0)

		// Get performance stats
		stats := fallbackMgr.GetPerformanceStats()
		if stats == nil {
			t.Fatal("Failed to get performance stats")
		}

		// Check CPU stats
		if stats.CPUStats.Count != 3 {
			t.Errorf("Expected 3 CPU measurements, got %d", stats.CPUStats.Count)
		}
		if stats.CPUStats.AverageTime < 40 || stats.CPUStats.AverageTime > 60 {
			t.Errorf("Unexpected CPU average: %f", stats.CPUStats.AverageTime)
		}

		// Check GPU stats
		if stats.GPUStats.Count != 3 {
			t.Errorf("Expected 3 GPU measurements, got %d", stats.GPUStats.Count)
		}
		if stats.GPUStats.AverageTime < 15 || stats.GPUStats.AverageTime > 30 {
			t.Errorf("Unexpected GPU average: %f", stats.GPUStats.AverageTime)
		}

		t.Log("Performance monitoring test passed")
	})
}

// TestGPUPerformanceWithSimulation tests GPU performance in actual simulation
func TestGPUPerformanceWithSimulation(t *testing.T) {
	cfg := config.DefaultConfig()

	// Use different particle counts to test scaling
	particleCounts := []int{10, 50, 100}

	for _, numParticles := range particleCounts {
		t.Run(string(rune(numParticles))+"particles", func(t *testing.T) {
			cfg.NumParticles = numParticles

			particles := physics.InitializeParticles(
				cfg.NumParticles,
				float64(cfg.SimulationWidth),
				float64(cfg.SimulationDepth),
			)

			deltaTime := float32(0.01)
			iterations := 10

			// Warm up
			for i := 0; i < 5; i++ {
				physics.RunTimeEvolution(
					particles, deltaTime,
					cfg.SimulationWidth, cfg.SimulationDepth,
					cfg.GravitationalConstant,
				)
			}

			// Measure performance
			start := time.Now()
			for i := 0; i < iterations; i++ {
				physics.RunTimeEvolution(
					particles, deltaTime,
					cfg.SimulationWidth, cfg.SimulationDepth,
					cfg.GravitationalConstant,
				)
			}
			elapsed := time.Since(start)

			avgTime := elapsed / time.Duration(iterations)
			t.Logf("%d particles: %v per iteration", numParticles, avgTime)

			// Performance should be reasonable
			maxTime := 200 * time.Millisecond
			if avgTime > maxTime {
				t.Errorf("Performance issue: %v per iteration (expected < %v)",
					avgTime, maxTime)
			}
		})
	}
}

// TestBufferPooling tests buffer pooling for performance
func TestBufferPooling(t *testing.T) {
	bufferMgr := gpu.NewBufferManager()
	if bufferMgr == nil {
		t.Fatal("Failed to create buffer manager")
	}

	// Test getting pooled buffers
	size := 1024
	buffer1 := bufferMgr.GetPooledBuffer(size)
	if buffer1 == nil {
		// Pooled buffer may not exist initially
		t.Log("No pooled buffer available (expected on first request)")

		// Create a buffer and return it to pool
		buffer, err := bufferMgr.CreateFloatBuffer(size)
		if err != nil {
			t.Errorf("Failed to create buffer: %v", err)
			return
		}

		bufferMgr.ReturnToPool(buffer)

		// Now try to get from pool again
		buffer2 := bufferMgr.GetPooledBuffer(size)
		if buffer2 != nil {
			t.Log("Successfully retrieved buffer from pool")
			_ = bufferMgr.FreeBuffer(buffer2)
		}
	} else {
		// Return the buffer for reuse
		bufferMgr.ReturnToPool(buffer1)

		// Get it again
		buffer2 := bufferMgr.GetPooledBuffer(size)
		if buffer2 == nil {
			t.Error("Failed to retrieve buffer from pool")
		} else {
			t.Log("Buffer pooling working correctly")
			_ = bufferMgr.FreeBuffer(buffer2)
		}
	}
}

// BenchmarkBufferTransfer benchmarks data transfer to/from GPU
func BenchmarkBufferTransfer(b *testing.B) {
	bufferMgr := gpu.NewBufferManager()
	if bufferMgr == nil {
		b.Fatal("Failed to create buffer manager")
	}

	size := 1024 * 1024 // 1M floats
	buffer, err := bufferMgr.CreateFloatBuffer(size)
	if err != nil {
		b.Fatalf("Failed to create buffer: %v", err)
	}
	defer func() {
		_ = bufferMgr.FreeBuffer(buffer)
	}()

	data := make([]float32, size)
	for i := range data {
		data[i] = float32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Upload
		err := bufferMgr.UploadFloatData(buffer, data)
		if err != nil {
			b.Fatalf("Upload failed: %v", err)
		}

		// Download
		_, err = bufferMgr.DownloadFloatData(buffer, size)
		if err != nil {
			b.Fatalf("Download failed: %v", err)
		}
	}
}
