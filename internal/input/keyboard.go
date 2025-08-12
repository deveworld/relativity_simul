package input

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Movement represents movement input in 3D space
type Movement struct {
	Forward float32
	Right   float32
	Up      float32
}

// Actions represents action inputs from keyboard
type Actions struct {
	TogglePause bool
	ToggleGPU   bool
}

// KeyboardHandler handles keyboard input
type KeyboardHandler struct {
	keyStates  map[int32]bool
	keyPressed map[int32]bool
}

// NewKeyboardHandler creates a new keyboard handler
func NewKeyboardHandler() *KeyboardHandler {
	return &KeyboardHandler{
		keyStates:  make(map[int32]bool),
		keyPressed: make(map[int32]bool),
	}
}

// SetKeyState sets the state of a key (for testing)
func (k *KeyboardHandler) SetKeyState(key int32, pressed bool) {
	k.keyStates[key] = pressed
}

// SetKeyPressed sets whether a key was just pressed (for testing)
func (k *KeyboardHandler) SetKeyPressed(key int32, pressed bool) {
	k.keyPressed[key] = pressed
}

// IsKeyDown checks if a key is currently held down
func (k *KeyboardHandler) IsKeyDown(key int32) bool {
	// In real usage, this would call rl.IsKeyDown(key)
	// For testing, we use our map
	return k.keyStates[key]
}

// IsKeyPressed checks if a key was just pressed
func (k *KeyboardHandler) IsKeyPressed(key int32) bool {
	// In real usage, this would call rl.IsKeyPressed(key)
	// For testing, we use our map
	return k.keyPressed[key]
}

// ProcessMovement processes movement keys and returns movement deltas
func (k *KeyboardHandler) ProcessMovement(yaw, moveSpeed float32) *Movement {
	movement := &Movement{}

	// Calculate forward and right vectors based on yaw
	forward := rl.NewVector3(
		float32(math.Cos(float64(yaw))),
		0,
		float32(math.Sin(float64(yaw))),
	)
	right := rl.NewVector3(
		float32(math.Cos(float64(yaw-1.5708))), // yaw - π/2
		0,
		float32(math.Sin(float64(yaw-1.5708))),
	)

	// Process movement keys
	if k.IsKeyDown(rl.KeyW) {
		movement.Forward += moveSpeed
	}
	if k.IsKeyDown(rl.KeyS) {
		movement.Forward -= moveSpeed
	}
	if k.IsKeyDown(rl.KeyA) {
		movement.Right -= moveSpeed
	}
	if k.IsKeyDown(rl.KeyD) {
		movement.Right += moveSpeed
	}
	if k.IsKeyDown(rl.KeyQ) {
		movement.Up -= moveSpeed
	}
	if k.IsKeyDown(rl.KeyE) {
		movement.Up += moveSpeed
	}

	// Apply direction vectors (simplified for test)
	_ = forward
	_ = right

	return movement
}

// ProcessActions processes action keys and returns action flags
func (k *KeyboardHandler) ProcessActions() *Actions {
	return &Actions{
		TogglePause: k.IsKeyPressed(rl.KeyP),
		ToggleGPU:   k.IsKeyPressed(rl.KeyG),
	}
}

// UpdateFromRaylib updates key states from raylib (for production use)
func (k *KeyboardHandler) UpdateFromRaylib() {
	// Clear pressed states each frame
	k.keyPressed = make(map[int32]bool)

	// Update key pressed states
	k.keyPressed[rl.KeyP] = rl.IsKeyPressed(rl.KeyP)
	k.keyPressed[rl.KeyG] = rl.IsKeyPressed(rl.KeyG)

	// Update key held states
	k.keyStates[rl.KeyW] = rl.IsKeyDown(rl.KeyW)
	k.keyStates[rl.KeyS] = rl.IsKeyDown(rl.KeyS)
	k.keyStates[rl.KeyA] = rl.IsKeyDown(rl.KeyA)
	k.keyStates[rl.KeyD] = rl.IsKeyDown(rl.KeyD)
	k.keyStates[rl.KeyQ] = rl.IsKeyDown(rl.KeyQ)
	k.keyStates[rl.KeyE] = rl.IsKeyDown(rl.KeyE)
}

// ProcessKeyboardInput processes keyboard input for camera movement
func ProcessKeyboardInput(camera *rl.Camera3D, yaw, moveSpeed float32, pause *bool, useGPU *bool) {
	handler := NewKeyboardHandler()
	handler.UpdateFromRaylib()

	// Process actions
	actions := handler.ProcessActions()
	if actions.TogglePause {
		*pause = !*pause
	}
	if actions.ToggleGPU {
		*useGPU = !*useGPU
	}

	// Process movement
	movement := handler.ProcessMovement(yaw, moveSpeed)

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
