package input

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKeyboardInput is a mock for testing keyboard input
type MockKeyboardInput struct {
	mock.Mock
}

func (m *MockKeyboardInput) IsKeyDown(key int32) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockKeyboardInput) IsKeyPressed(key int32) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func TestKeyboardHandler_ProcessMovement(t *testing.T) {
	handler := NewKeyboardHandler()

	// Test W key for forward movement
	t.Run("W key moves forward", func(t *testing.T) {
		movement := handler.ProcessMovement(0.0, 0.0) // yaw and moveSpeed
		assert.NotNil(t, movement)

		// Simulate W key pressed
		handler.SetKeyState(rl.KeyW, true)
		movement = handler.ProcessMovement(0.0, 1.0)

		// Forward movement should be positive Z
		assert.Greater(t, movement.Forward, float32(0.0))
		assert.Equal(t, float32(0.0), movement.Right)
		assert.Equal(t, float32(0.0), movement.Up)
	})

	// Test S key for backward movement
	t.Run("S key moves backward", func(t *testing.T) {
		handler := NewKeyboardHandler()
		handler.SetKeyState(rl.KeyS, true)
		movement := handler.ProcessMovement(0.0, 1.0)

		assert.Less(t, movement.Forward, float32(0.0))
		assert.Equal(t, float32(0.0), movement.Right)
		assert.Equal(t, float32(0.0), movement.Up)
	})

	// Test A key for left movement
	t.Run("A key moves left", func(t *testing.T) {
		handler := NewKeyboardHandler()
		handler.SetKeyState(rl.KeyA, true)
		movement := handler.ProcessMovement(0.0, 1.0)

		assert.Equal(t, float32(0.0), movement.Forward)
		assert.Less(t, movement.Right, float32(0.0)) // Left is negative right
		assert.Equal(t, float32(0.0), movement.Up)
	})

	// Test D key for right movement
	t.Run("D key moves right", func(t *testing.T) {
		handler := NewKeyboardHandler()
		handler.SetKeyState(rl.KeyD, true)
		movement := handler.ProcessMovement(0.0, 1.0)

		assert.Equal(t, float32(0.0), movement.Forward)
		assert.Greater(t, movement.Right, float32(0.0)) // Right is positive right
		assert.Equal(t, float32(0.0), movement.Up)
	})

	// Test Q key for down movement
	t.Run("Q key moves down", func(t *testing.T) {
		handler := NewKeyboardHandler()
		handler.SetKeyState(rl.KeyQ, true)
		movement := handler.ProcessMovement(0.0, 1.0)

		assert.Equal(t, float32(0.0), movement.Forward)
		assert.Equal(t, float32(0.0), movement.Right)
		assert.Less(t, movement.Up, float32(0.0))
	})

	// Test E key for up movement
	t.Run("E key moves up", func(t *testing.T) {
		handler := NewKeyboardHandler()
		handler.SetKeyState(rl.KeyE, true)
		movement := handler.ProcessMovement(0.0, 1.0)

		assert.Equal(t, float32(0.0), movement.Forward)
		assert.Equal(t, float32(0.0), movement.Right)
		assert.Greater(t, movement.Up, float32(0.0))
	})
}

func TestKeyboardHandler_ProcessActions(t *testing.T) {
	handler := NewKeyboardHandler()

	// Test P key for pause toggle
	t.Run("P key toggles pause", func(t *testing.T) {
		actions := handler.ProcessActions()
		assert.False(t, actions.TogglePause)

		handler.SetKeyPressed(rl.KeyP, true)
		actions = handler.ProcessActions()
		assert.True(t, actions.TogglePause)

		// Reset and check it's false again
		handler.SetKeyPressed(rl.KeyP, false)
		actions = handler.ProcessActions()
		assert.False(t, actions.TogglePause)
	})

	// Test G key for GPU toggle
	t.Run("G key toggles GPU", func(t *testing.T) {
		handler := NewKeyboardHandler()
		actions := handler.ProcessActions()
		assert.False(t, actions.ToggleGPU)

		handler.SetKeyPressed(rl.KeyG, true)
		actions = handler.ProcessActions()
		assert.True(t, actions.ToggleGPU)

		// Reset and check it's false again
		handler.SetKeyPressed(rl.KeyG, false)
		actions = handler.ProcessActions()
		assert.False(t, actions.ToggleGPU)
	})
}

func TestKeyboardHandler_CombinedMovement(t *testing.T) {
	handler := NewKeyboardHandler()

	// Test combined W+D movement (forward-right diagonal)
	t.Run("W+D moves forward-right", func(t *testing.T) {
		handler.SetKeyState(rl.KeyW, true)
		handler.SetKeyState(rl.KeyD, true)
		movement := handler.ProcessMovement(0.0, 1.0)

		assert.Greater(t, movement.Forward, float32(0.0))
		assert.Greater(t, movement.Right, float32(0.0)) // D moves right (positive)
		assert.Equal(t, float32(0.0), movement.Up)
	})

	// Test all movement keys at once
	t.Run("Multiple keys combine correctly", func(t *testing.T) {
		handler := NewKeyboardHandler()
		handler.SetKeyState(rl.KeyW, true)
		handler.SetKeyState(rl.KeyS, true) // Cancel each other out
		handler.SetKeyState(rl.KeyA, true)
		handler.SetKeyState(rl.KeyD, true) // Cancel each other out
		handler.SetKeyState(rl.KeyQ, true)
		handler.SetKeyState(rl.KeyE, true) // Cancel each other out

		movement := handler.ProcessMovement(0.0, 1.0)

		// All movements should cancel out
		assert.InDelta(t, 0.0, float64(movement.Forward), 0.001)
		assert.InDelta(t, 0.0, float64(movement.Right), 0.001)
		assert.InDelta(t, 0.0, float64(movement.Up), 0.001)
	})
}
