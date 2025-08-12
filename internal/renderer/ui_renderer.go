package renderer

import (
	"errors"
	"fmt"
)

// ComputeMode represents the compute mode for UI display
type ComputeMode int

const (
	// ModeCPU represents CPU compute mode
	ModeCPU ComputeMode = iota
	// ModeGPU represents GPU compute mode
	ModeGPU
)

// UIColor represents an RGB color for UI elements
type UIColor struct {
	R, G, B, A uint8
}

// UIState represents the current UI state
type UIState struct {
	ParticleCount int
	Mode          ComputeMode
	GPUFallback   bool
	TargetFPS     int
	ActualFPS     int
	FrameTime     float64
	Paused        bool
}

// UIRenderer handles UI rendering
type UIRenderer struct {
	screenWidth  int
	screenHeight int
	fontSize     int

	// UI state
	title         string
	particleCount int
	mode          ComputeMode
	gpuFallback   bool
	targetFPS     int
	actualFPS     int
	frameTime     float64
	paused        bool
}

// NewUIRenderer creates a new UI renderer
func NewUIRenderer(screenWidth, screenHeight int) *UIRenderer {
	return &UIRenderer{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		fontSize:     20, // Default font size from main.go
		title:        "GR (Weak-Field) N-Body Simulation",
	}
}

// GetScreenDimensions returns the screen dimensions
func (ui *UIRenderer) GetScreenDimensions() (int, int) {
	return ui.screenWidth, ui.screenHeight
}

// SetTitle sets the UI title
func (ui *UIRenderer) SetTitle(title string) {
	ui.title = title
}

// GetTitle returns the UI title
func (ui *UIRenderer) GetTitle() string {
	return ui.title
}

// SetParticleCount sets the particle count
func (ui *UIRenderer) SetParticleCount(count int) {
	ui.particleCount = count
}

// GetParticleCount returns the particle count
func (ui *UIRenderer) GetParticleCount() int {
	return ui.particleCount
}

// SetMode sets the compute mode
func (ui *UIRenderer) SetMode(mode ComputeMode, fallback bool) {
	ui.mode = mode
	ui.gpuFallback = fallback
}

// GetModeString returns the mode display string
func (ui *UIRenderer) GetModeString() string {
	switch ui.mode {
	case ModeGPU:
		if ui.gpuFallback {
			return "Mode: GPU (Fallback to CPU)"
		}
		return "Mode: GPU Accelerated"
	case ModeCPU:
		return "Mode: CPU Only"
	default:
		return "Mode: Unknown"
	}
}

// GetControlInstructions returns the control instruction lines
func (ui *UIRenderer) GetControlInstructions() []string {
	return []string{
		"Right-click + Mouse to look",
		"W,A,S,D,Q,E to move",
		"P to pause, G to toggle GPU",
	}
}

// SetTargetFPS sets the target FPS
func (ui *UIRenderer) SetTargetFPS(fps int) {
	ui.targetFPS = fps
}

// GetTargetFPS returns the target FPS
func (ui *UIRenderer) GetTargetFPS() int {
	return ui.targetFPS
}

// SetActualFPS sets the actual FPS
func (ui *UIRenderer) SetActualFPS(fps int) {
	ui.actualFPS = fps
}

// GetActualFPS returns the actual FPS
func (ui *UIRenderer) GetActualFPS() int {
	return ui.actualFPS
}

// SetFrameTime sets the frame time
func (ui *UIRenderer) SetFrameTime(time float64) {
	ui.frameTime = time
}

// GetFrameTime returns the frame time
func (ui *UIRenderer) GetFrameTime() float64 {
	return ui.frameTime
}

// SetPaused sets the pause state
func (ui *UIRenderer) SetPaused(paused bool) {
	ui.paused = paused
}

// IsPaused returns the pause state
func (ui *UIRenderer) IsPaused() bool {
	return ui.paused
}

// GetPauseText returns the pause indicator text
func (ui *UIRenderer) GetPauseText() string {
	return "PAUSED (Press P to unpause)"
}

// GetTitlePosition returns the title position
func (ui *UIRenderer) GetTitlePosition() (int, int) {
	return 10, 10
}

// GetParticleCountPosition returns the particle count position
func (ui *UIRenderer) GetParticleCountPosition() (int, int) {
	return 10, 40
}

// GetModePosition returns the mode display position
func (ui *UIRenderer) GetModePosition() (int, int) {
	return 10, 70
}

// GetFPSPosition returns the FPS display position
func (ui *UIRenderer) GetFPSPosition() (int, int) {
	return ui.screenWidth - 200, 10
}

// GetPausePosition returns the pause indicator position
func (ui *UIRenderer) GetPausePosition() (int, int) {
	return ui.screenWidth/2 - 150, ui.screenHeight/2 - 10
}

// GetTitleColor returns the title color (lime/green)
func (ui *UIRenderer) GetTitleColor() UIColor {
	return UIColor{R: 0, G: 255, B: 0, A: 255}
}

// GetDefaultTextColor returns the default text color (white)
func (ui *UIRenderer) GetDefaultTextColor() UIColor {
	return UIColor{R: 255, G: 255, B: 255, A: 255}
}

// GetModeColor returns the color for mode display
func (ui *UIRenderer) GetModeColor(mode ComputeMode, fallback bool) UIColor {
	switch mode {
	case ModeGPU:
		if fallback {
			// Yellow for fallback
			return UIColor{R: 255, G: 255, B: 0, A: 255}
		}
		// Green for GPU
		return UIColor{R: 0, G: 255, B: 0, A: 255}
	case ModeCPU:
		// Orange for CPU
		return UIColor{R: 255, G: 165, B: 0, A: 255}
	default:
		return ui.GetDefaultTextColor()
	}
}

// GetPauseColor returns the pause indicator color (yellow)
func (ui *UIRenderer) GetPauseColor() UIColor {
	return UIColor{R: 255, G: 255, B: 0, A: 255}
}

// GetFontSize returns the font size
func (ui *UIRenderer) GetFontSize() int {
	return ui.fontSize
}

// SetFontSize sets the font size
func (ui *UIRenderer) SetFontSize(size int) {
	ui.fontSize = size
}

// UpdateState updates the UI state from a UIState struct
func (ui *UIRenderer) UpdateState(state UIState) {
	ui.particleCount = state.ParticleCount
	ui.mode = state.Mode
	ui.gpuFallback = state.GPUFallback
	ui.targetFPS = state.TargetFPS
	ui.actualFPS = state.ActualFPS
	ui.frameTime = state.FrameTime
	ui.paused = state.Paused
}

// Render renders the UI (mock implementation)
func (ui *UIRenderer) Render() error {
	// In a real implementation, this would draw all UI elements
	// For now, return an error since we don't have a graphics context
	return errors.New("graphics context not available")
}

// GetParticleCountText returns formatted particle count text
func (ui *UIRenderer) GetParticleCountText() string {
	return fmt.Sprintf("Particles: %d", ui.particleCount)
}

// GetTargetFPSText returns formatted target FPS text
func (ui *UIRenderer) GetTargetFPSText() string {
	return fmt.Sprintf("Target FPS: %d", ui.targetFPS)
}

// GetActualFPSText returns formatted actual FPS text
func (ui *UIRenderer) GetActualFPSText() string {
	return fmt.Sprintf("Actual FPS: %d", ui.actualFPS)
}

// GetFrameTimeText returns formatted frame time text
func (ui *UIRenderer) GetFrameTimeText() string {
	return fmt.Sprintf("Frame Time: %.3fs", ui.frameTime)
}

// GetControlPosition returns the position for control instruction at given index
func (ui *UIRenderer) GetControlPosition(index int) (int, int) {
	// Control instructions start at y=130 with 30 pixel spacing
	return 10, 130 + index*30
}

// GetActualFPSPosition returns the actual FPS display position
func (ui *UIRenderer) GetActualFPSPosition() (int, int) {
	return ui.screenWidth - 200, 35
}

// GetFrameTimePosition returns the frame time display position
func (ui *UIRenderer) GetFrameTimePosition() (int, int) {
	return ui.screenWidth - 200, 60
}
