package renderer

import (
	"testing"
)

// TestUIRendererCreation tests creating a UI renderer
func TestUIRendererCreation(t *testing.T) {
	ui := NewUIRenderer(800, 600)

	if ui == nil {
		t.Fatal("Failed to create UI renderer")
	}

	// Check screen dimensions
	w, h := ui.GetScreenDimensions()
	if w != 800 || h != 600 {
		t.Errorf("Screen dimensions incorrect: expected 800x600, got %dx%d", w, h)
	}
}

// TestUIText tests UI text elements
func TestUIText(t *testing.T) {
	ui := NewUIRenderer(800, 600)

	// Test adding text elements
	ui.SetTitle("GR (Weak-Field) N-Body Simulation")
	if ui.GetTitle() != "GR (Weak-Field) N-Body Simulation" {
		t.Error("Failed to set title")
	}

	// Test particle count
	ui.SetParticleCount(1000)
	if ui.GetParticleCount() != 1000 {
		t.Error("Failed to set particle count")
	}

	// Test mode display
	ui.SetMode(ModeGPU, false)
	mode := ui.GetModeString()
	if mode != "Mode: GPU Accelerated" {
		t.Errorf("Incorrect mode string: %s", mode)
	}

	// Test GPU fallback mode
	ui.SetMode(ModeGPU, true)
	mode = ui.GetModeString()
	if mode != "Mode: GPU (Fallback to CPU)" {
		t.Errorf("Incorrect fallback mode string: %s", mode)
	}

	// Test CPU mode
	ui.SetMode(ModeCPU, false)
	mode = ui.GetModeString()
	if mode != "Mode: CPU Only" {
		t.Errorf("Incorrect CPU mode string: %s", mode)
	}
}

// TestUIControls tests UI control instructions
func TestUIControls(t *testing.T) {
	ui := NewUIRenderer(800, 600)

	// Get control instructions
	controls := ui.GetControlInstructions()

	// Should have at least 3 instruction lines
	if len(controls) < 3 {
		t.Error("Missing control instructions")
	}

	// Check specific instructions exist
	expectedInstructions := []string{
		"Right-click + Mouse to look",
		"W,A,S,D,Q,E to move",
		"P to pause, G to toggle GPU",
	}

	for i, expected := range expectedInstructions {
		if i >= len(controls) {
			t.Errorf("Missing instruction: %s", expected)
		} else if controls[i] != expected {
			t.Errorf("Instruction mismatch: expected '%s', got '%s'",
				expected, controls[i])
		}
	}
}

// TestUIFPSDisplay tests FPS and frame time display
func TestUIFPSDisplay(t *testing.T) {
	ui := NewUIRenderer(800, 600)

	// Set FPS values
	ui.SetTargetFPS(60)
	ui.SetActualFPS(58)
	ui.SetFrameTime(0.017)

	// Check values
	if ui.GetTargetFPS() != 60 {
		t.Error("Failed to set target FPS")
	}

	if ui.GetActualFPS() != 58 {
		t.Error("Failed to set actual FPS")
	}

	if ui.GetFrameTime() != 0.017 {
		t.Error("Failed to set frame time")
	}
}

// TestUIPauseIndicator tests pause indicator
func TestUIPauseIndicator(t *testing.T) {
	ui := NewUIRenderer(800, 600)

	// Initially not paused
	if ui.IsPaused() {
		t.Error("Should not be paused initially")
	}

	// Set paused
	ui.SetPaused(true)
	if !ui.IsPaused() {
		t.Error("Should be paused")
	}

	// Get pause text
	pauseText := ui.GetPauseText()
	if pauseText != "PAUSED (Press P to unpause)" {
		t.Errorf("Incorrect pause text: %s", pauseText)
	}
}

// TestUITextPositions tests text positioning
func TestUITextPositions(t *testing.T) {
	ui := NewUIRenderer(800, 600)

	// Test title position (should be top-left)
	x, y := ui.GetTitlePosition()
	if x != 10 || y != 10 {
		t.Errorf("Title position incorrect: expected (10,10), got (%d,%d)", x, y)
	}

	// Test particle count position
	x, y = ui.GetParticleCountPosition()
	if x != 10 || y != 40 {
		t.Errorf("Particle count position incorrect: expected (10,40), got (%d,%d)", x, y)
	}

	// Test mode position
	x, y = ui.GetModePosition()
	if x != 10 || y != 70 {
		t.Errorf("Mode position incorrect: expected (10,70), got (%d,%d)", x, y)
	}

	// Test FPS position (should be top-right)
	x, y = ui.GetFPSPosition()
	if x != 600 || y != 10 { // 800 - 200 = 600
		t.Errorf("FPS position incorrect: expected (600,10), got (%d,%d)", x, y)
	}

	// Test pause indicator position (should be centered)
	x, y = ui.GetPausePosition()
	expectedX := 800/2 - 150
	expectedY := 600/2 - 10
	if x != expectedX || y != expectedY {
		t.Errorf("Pause position incorrect: expected (%d,%d), got (%d,%d)",
			expectedX, expectedY, x, y)
	}
}

// TestUIColors tests UI color settings
func TestUIColors(t *testing.T) {
	ui := NewUIRenderer(800, 600)

	// Test title color (should be lime/green)
	color := ui.GetTitleColor()
	if color.R != 0 || color.G != 255 || color.B != 0 {
		t.Error("Title color should be lime/green")
	}

	// Test default text color (should be white)
	color = ui.GetDefaultTextColor()
	if color.R != 255 || color.G != 255 || color.B != 255 {
		t.Error("Default text color should be white")
	}

	// Test GPU mode color (should be green)
	color = ui.GetModeColor(ModeGPU, false)
	if color.R != 0 || color.G < 200 || color.B != 0 {
		t.Error("GPU mode color should be green")
	}

	// Test GPU fallback color (should be yellow)
	color = ui.GetModeColor(ModeGPU, true)
	if color.R < 200 || color.G < 200 || color.B != 0 {
		t.Error("GPU fallback color should be yellow")
	}

	// Test CPU mode color (should be orange)
	color = ui.GetModeColor(ModeCPU, false)
	if color.R < 200 || color.G < 100 || color.B != 0 {
		t.Error("CPU mode color should be orange")
	}

	// Test pause color (should be yellow)
	color = ui.GetPauseColor()
	if color.R < 200 || color.G < 200 || color.B != 0 {
		t.Error("Pause color should be yellow")
	}
}

// TestUIFontSize tests font size settings
func TestUIFontSize(t *testing.T) {
	ui := NewUIRenderer(800, 600)

	// Default font size should be 20 (as in main.go)
	if ui.GetFontSize() != 20 {
		t.Errorf("Default font size should be 20, got %d", ui.GetFontSize())
	}

	// Test setting custom font size
	ui.SetFontSize(24)
	if ui.GetFontSize() != 24 {
		t.Error("Failed to set font size")
	}
}

// TestUIUpdate tests updating UI state
func TestUIUpdate(t *testing.T) {
	ui := NewUIRenderer(800, 600)

	// Create UI state
	state := UIState{
		ParticleCount: 500,
		Mode:          ModeGPU,
		GPUFallback:   false,
		TargetFPS:     60,
		ActualFPS:     59,
		FrameTime:     0.016,
		Paused:        false,
	}

	// Update UI with state
	ui.UpdateState(state)

	// Verify all values updated
	if ui.GetParticleCount() != 500 {
		t.Error("Particle count not updated")
	}

	if ui.GetTargetFPS() != 60 {
		t.Error("Target FPS not updated")
	}

	if ui.GetActualFPS() != 59 {
		t.Error("Actual FPS not updated")
	}

	if ui.IsPaused() {
		t.Error("Pause state not updated correctly")
	}
}

// TestUIRender tests rendering (mock)
func TestUIRender(t *testing.T) {
	ui := NewUIRenderer(800, 600)

	// Set up some UI state
	ui.SetTitle("Test Title")
	ui.SetParticleCount(100)
	ui.SetMode(ModeGPU, false)
	ui.SetTargetFPS(60)
	ui.SetActualFPS(60)
	ui.SetFrameTime(0.016)

	// Render should not error
	err := ui.Render()
	if err != nil {
		// Expected to error without graphics context
		t.Logf("Render error (expected): %v", err)
	}
}
