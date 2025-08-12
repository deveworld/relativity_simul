package input

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Rotation represents mouse rotation state
type Rotation struct {
	Active       bool
	YawDelta     float32
	PitchDelta   float32
	ShouldCenter bool
}

// MouseHandler handles mouse input
type MouseHandler struct {
	buttonStates map[rl.MouseButton]bool
	deltaX       float32
	deltaY       float32
}

// NewMouseHandler creates a new mouse handler
func NewMouseHandler() *MouseHandler {
	return &MouseHandler{
		buttonStates: make(map[rl.MouseButton]bool),
	}
}

// SetButtonDown sets the state of a mouse button (for testing)
func (m *MouseHandler) SetButtonDown(button rl.MouseButton, down bool) {
	m.buttonStates[button] = down
}

// SetMouseDelta sets the mouse delta (for testing)
func (m *MouseHandler) SetMouseDelta(x, y float32) {
	m.deltaX = x
	m.deltaY = y
}

// IsButtonDown checks if a mouse button is held down
func (m *MouseHandler) IsButtonDown(button rl.MouseButton) bool {
	// In real usage, this would call rl.IsMouseButtonDown(button)
	// For testing, we use our map
	return m.buttonStates[button]
}

// GetMouseDelta gets the mouse movement delta
func (m *MouseHandler) GetMouseDelta() (float32, float32) {
	// In real usage, this would call rl.GetMouseDelta()
	// For testing, we use our stored values
	return m.deltaX, m.deltaY
}

// ProcessRotation processes mouse rotation input
func (m *MouseHandler) ProcessRotation(currentYaw, currentPitch, sensitivity float32) *Rotation {
	rotation := &Rotation{
		Active:       false,
		YawDelta:     0,
		PitchDelta:   0,
		ShouldCenter: false,
	}

	if !m.IsButtonDown(rl.MouseRightButton) {
		rotation.ShouldCenter = true
		return rotation
	}

	rotation.Active = true
	deltaX, deltaY := m.GetMouseDelta()

	// Calculate rotation deltas
	rotation.YawDelta = deltaX * sensitivity
	rotation.PitchDelta = -deltaY * sensitivity

	// Clamp pitch to prevent flipping
	newPitch := currentPitch + rotation.PitchDelta
	if newPitch > 1.5 {
		rotation.PitchDelta = 1.5 - currentPitch
	} else if newPitch < -1.5 {
		rotation.PitchDelta = -1.5 - currentPitch
	}

	return rotation
}

// UpdateCameraTarget updates the camera target based on yaw and pitch
func (m *MouseHandler) UpdateCameraTarget(camera *rl.Camera3D, yaw, pitch float32) {
	camera.Target.X = camera.Position.X + float32(math.Cos(float64(yaw))*math.Cos(float64(pitch)))
	camera.Target.Y = camera.Position.Y + float32(math.Sin(float64(pitch)))
	camera.Target.Z = camera.Position.Z + float32(math.Sin(float64(yaw))*math.Cos(float64(pitch)))
}

// UpdateFromRaylib updates mouse state from raylib (for production use)
func (m *MouseHandler) UpdateFromRaylib() {
	// Update button states
	m.buttonStates[rl.MouseRightButton] = rl.IsMouseButtonDown(rl.MouseRightButton)
	m.buttonStates[rl.MouseLeftButton] = rl.IsMouseButtonDown(rl.MouseLeftButton)

	// Update delta
	delta := rl.GetMouseDelta()
	m.deltaX = delta.X
	m.deltaY = delta.Y
}

// ProcessMouseInput processes mouse input for camera rotation
func ProcessMouseInput(camera *rl.Camera3D, yaw, pitch *float32, mouseSensitivity float32, screenWidth, screenHeight int) {
	handler := NewMouseHandler()
	handler.UpdateFromRaylib()

	rotation := handler.ProcessRotation(*yaw, *pitch, mouseSensitivity)

	if rotation.ShouldCenter {
		rl.SetMousePosition(screenWidth/2, screenHeight/2)
	} else if rotation.Active {
		*yaw += rotation.YawDelta
		*pitch += rotation.PitchDelta
		handler.UpdateCameraTarget(camera, *yaw, *pitch)
	}
}
