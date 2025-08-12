package input

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// SimulationState holds the current simulation state affected by input
type SimulationState struct {
	Pause  bool
	UseGPU bool
	Yaw    float32
	Pitch  float32
}

// InputConfig holds input configuration settings
type InputConfig struct {
	MoveSpeed        float32
	MouseSensitivity float32
	ScreenWidth      int
	ScreenHeight     int
}

// InputController coordinates keyboard and mouse input
type InputController struct {
	keyboard *KeyboardHandler
	mouse    *MouseHandler
}

// NewInputController creates a new input controller
func NewInputController() *InputController {
	return &InputController{
		keyboard: NewKeyboardHandler(),
		mouse:    NewMouseHandler(),
	}
}

// ProcessInput processes all input and updates camera and state
func (c *InputController) ProcessInput(camera *rl.Camera3D, state *SimulationState, config *InputConfig) {
	// Process keyboard actions
	actions := c.keyboard.ProcessActions()
	if actions.TogglePause {
		state.Pause = !state.Pause
	}
	if actions.ToggleGPU {
		state.UseGPU = !state.UseGPU
	}

	// Process keyboard movement
	movement := c.keyboard.ProcessMovement(state.Yaw, config.MoveSpeed)
	applyMovement(camera, movement, state.Yaw)

	// Process mouse rotation
	rotation := c.mouse.ProcessRotation(state.Yaw, state.Pitch, config.MouseSensitivity)
	if rotation.ShouldCenter {
		rl.SetMousePosition(config.ScreenWidth/2, config.ScreenHeight/2)
	} else if rotation.Active {
		state.Yaw += rotation.YawDelta
		state.Pitch += rotation.PitchDelta
		c.mouse.UpdateCameraTarget(camera, state.Yaw, state.Pitch)
	}
}

// UpdateFromRaylib updates input states from raylib
func (c *InputController) UpdateFromRaylib() {
	c.keyboard.UpdateFromRaylib()
	c.mouse.UpdateFromRaylib()
}

// Reset clears all input states
func (c *InputController) Reset() {
	c.keyboard.keyStates = make(map[int32]bool)
	c.keyboard.keyPressed = make(map[int32]bool)
	c.mouse.buttonStates = make(map[rl.MouseButton]bool)
	c.mouse.deltaX = 0
	c.mouse.deltaY = 0
}

// applyMovement applies movement to the camera
func applyMovement(camera *rl.Camera3D, movement *Movement, yaw float32) {
	// Calculate direction vectors
	forward := rl.NewVector3(
		float32(math.Cos(float64(yaw))),
		0,
		float32(math.Sin(float64(yaw))),
	)
	right := rl.NewVector3(
		float32(math.Cos(float64(yaw-1.5708))),
		0,
		float32(math.Sin(float64(yaw-1.5708))),
	)

	// Apply forward/backward movement
	if movement.Forward != 0 {
		camera.Position.X += forward.X * movement.Forward
		camera.Position.Z += forward.Z * movement.Forward
		camera.Target.X += forward.X * movement.Forward
		camera.Target.Z += forward.Z * movement.Forward
	}

	// Apply left/right movement
	if movement.Right != 0 {
		camera.Position.X -= right.X * movement.Right
		camera.Position.Z -= right.Z * movement.Right
		camera.Target.X -= right.X * movement.Right
		camera.Target.Z -= right.Z * movement.Right
	}

	// Apply up/down movement
	if movement.Up != 0 {
		camera.Position.Y += movement.Up
		camera.Target.Y += movement.Up
	}
}

// ProcessAllInput is a convenience function that creates a controller and processes input
func ProcessAllInput(camera *rl.Camera3D, pause, useGPU *bool, yaw, pitch *float32, moveSpeed, mouseSensitivity float32, screenWidth, screenHeight int) {
	controller := NewInputController()
	controller.UpdateFromRaylib()

	state := &SimulationState{
		Pause:  *pause,
		UseGPU: *useGPU,
		Yaw:    *yaw,
		Pitch:  *pitch,
	}

	config := &InputConfig{
		MoveSpeed:        moveSpeed,
		MouseSensitivity: mouseSensitivity,
		ScreenWidth:      screenWidth,
		ScreenHeight:     screenHeight,
	}

	controller.ProcessInput(camera, state, config)

	// Update external state
	*pause = state.Pause
	*useGPU = state.UseGPU
	*yaw = state.Yaw
	*pitch = state.Pitch
}
