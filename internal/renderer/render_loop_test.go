package renderer

import (
	"testing"
	"time"
)

// TestRenderLoopCreation tests creating a render loop
func TestRenderLoopCreation(t *testing.T) {
	loop := NewRenderLoop()

	if loop == nil {
		t.Fatal("Failed to create render loop")
	}

	// Check default target FPS (should be 60 from main.go)
	if loop.GetTargetFPS() != 60 {
		t.Errorf("Default target FPS should be 60, got %d", loop.GetTargetFPS())
	}
}

// TestFrameTiming tests frame timing calculation
func TestFrameTiming(t *testing.T) {
	loop := NewRenderLoop()

	// Set target FPS
	loop.SetTargetFPS(60)

	// Get target frame time
	targetFrameTime := loop.GetTargetFrameTime()
	expectedTime := 1.0 / 60.0

	if targetFrameTime != expectedTime {
		t.Errorf("Target frame time incorrect: expected %f, got %f",
			expectedTime, targetFrameTime)
	}

	// Test frame time recording
	loop.RecordFrameTime(0.016)
	if loop.GetLastFrameTime() != 0.016 {
		t.Error("Failed to record frame time")
	}

	// Test actual FPS calculation
	loop.RecordFrameTime(0.0167) // ~60 FPS
	actualFPS := loop.GetActualFPS()
	if actualFPS < 59 || actualFPS > 61 {
		t.Errorf("Actual FPS calculation incorrect: got %d", actualFPS)
	}
}

// TestRenderLoopState tests render loop state management
func TestRenderLoopState(t *testing.T) {
	loop := NewRenderLoop()

	// Initially should not be running
	if loop.IsRunning() {
		t.Error("Loop should not be running initially")
	}

	// Start the loop
	loop.Start()
	if !loop.IsRunning() {
		t.Error("Loop should be running after Start()")
	}

	// Stop the loop
	loop.Stop()
	if loop.IsRunning() {
		t.Error("Loop should not be running after Stop()")
	}
}

// TestFrameRateLimit tests frame rate limiting
func TestFrameRateLimit(t *testing.T) {
	loop := NewRenderLoop()
	loop.SetTargetFPS(30) // Lower FPS for testing

	// Simulate fast frame
	startTime := time.Now()
	loop.BeginFrame()

	// Simulate very quick frame (1ms)
	time.Sleep(1 * time.Millisecond)

	// End frame should wait to maintain target FPS
	loop.EndFrame()

	elapsed := time.Since(startTime)
	targetTime := time.Duration(1000/30) * time.Millisecond

	// Should have waited to maintain 30 FPS
	if elapsed < targetTime-5*time.Millisecond {
		t.Errorf("Frame rate limiting not working: elapsed %v, expected ~%v",
			elapsed, targetTime)
	}
}

// TestRenderCallback tests render callback functionality
func TestRenderCallback(t *testing.T) {
	loop := NewRenderLoop()

	callbackCalled := false
	renderCount := 0

	// Set render callback
	loop.SetRenderCallback(func(dt float64) {
		callbackCalled = true
		renderCount++
	})

	// Execute one frame
	loop.ExecuteFrame()

	if !callbackCalled {
		t.Error("Render callback not called")
	}

	if renderCount != 1 {
		t.Errorf("Expected 1 render call, got %d", renderCount)
	}
}

// TestUpdateCallback tests update callback functionality
func TestUpdateCallback(t *testing.T) {
	loop := NewRenderLoop()

	updateCalled := false
	deltaTimeReceived := float64(0)

	// Set update callback
	loop.SetUpdateCallback(func(dt float64) {
		updateCalled = true
		deltaTimeReceived = dt
	})

	// Execute frame with known delta time
	loop.RecordFrameTime(0.016)
	loop.ExecuteFrame()

	if !updateCalled {
		t.Error("Update callback not called")
	}

	if deltaTimeReceived != 0.016 {
		t.Errorf("Incorrect delta time: expected 0.016, got %f", deltaTimeReceived)
	}
}

// TestFrameStatistics tests frame statistics tracking
func TestFrameStatistics(t *testing.T) {
	loop := NewRenderLoop()

	// Record multiple frame times
	frameTimes := []float64{0.016, 0.017, 0.015, 0.018, 0.016}
	for _, ft := range frameTimes {
		loop.RecordFrameTime(ft)
	}

	// Get average frame time
	avgFrameTime := loop.GetAverageFrameTime()
	expectedAvg := (0.016 + 0.017 + 0.015 + 0.018 + 0.016) / 5.0

	if avgFrameTime < expectedAvg-0.001 || avgFrameTime > expectedAvg+0.001 {
		t.Errorf("Average frame time incorrect: expected ~%f, got %f",
			expectedAvg, avgFrameTime)
	}

	// Get frame count
	if loop.GetFrameCount() != 5 {
		t.Errorf("Frame count incorrect: expected 5, got %d", loop.GetFrameCount())
	}
}

// TestVSync tests vsync setting
func TestVSync(t *testing.T) {
	loop := NewRenderLoop()

	// Initially vsync should be off
	if loop.IsVSyncEnabled() {
		t.Error("VSync should be disabled by default")
	}

	// Enable vsync
	loop.EnableVSync(true)
	if !loop.IsVSyncEnabled() {
		t.Error("Failed to enable VSync")
	}

	// Disable vsync
	loop.EnableVSync(false)
	if loop.IsVSyncEnabled() {
		t.Error("Failed to disable VSync")
	}
}

// TestRenderPhases tests render phase management
func TestRenderPhases(t *testing.T) {
	loop := NewRenderLoop()

	phasesCalled := []string{}

	// Set callbacks for different phases
	loop.SetBeginCallback(func() {
		phasesCalled = append(phasesCalled, "begin")
	})

	loop.SetEndCallback(func() {
		phasesCalled = append(phasesCalled, "end")
	})

	loop.SetUpdateCallback(func(dt float64) {
		phasesCalled = append(phasesCalled, "update")
	})

	loop.SetRenderCallback(func(dt float64) {
		phasesCalled = append(phasesCalled, "render")
	})

	// Execute frame
	loop.ExecuteFrame()

	// Check phase order
	expectedOrder := []string{"begin", "update", "render", "end"}

	if len(phasesCalled) != len(expectedOrder) {
		t.Errorf("Expected %d phases, got %d", len(expectedOrder), len(phasesCalled))
	}

	for i, phase := range expectedOrder {
		if i >= len(phasesCalled) || phasesCalled[i] != phase {
			t.Errorf("Phase %d: expected '%s', got '%s'",
				i, phase, phasesCalled[i])
		}
	}
}

// TestShouldClose tests the should close functionality
func TestShouldClose(t *testing.T) {
	loop := NewRenderLoop()

	// Initially should not close
	if loop.ShouldClose() {
		t.Error("Should not close initially")
	}

	// Request close
	loop.RequestClose()
	if !loop.ShouldClose() {
		t.Error("Should close after request")
	}

	// Reset
	loop.ResetCloseRequest()
	if loop.ShouldClose() {
		t.Error("Should not close after reset")
	}
}
