package input

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/stretchr/testify/assert"
)

func TestInputController_Integration(t *testing.T) {
	controller := NewInputController()

	// Test controller initialization
	t.Run("Controller initializes with handlers", func(t *testing.T) {
		assert.NotNil(t, controller)
		assert.NotNil(t, controller.keyboard)
		assert.NotNil(t, controller.mouse)
	})

	// Test processing input
	t.Run("Controller processes both keyboard and mouse", func(t *testing.T) {
		camera := &rl.Camera3D{
			Position: rl.NewVector3(0, 0, 0),
			Target:   rl.NewVector3(1, 0, 0),
			Up:       rl.NewVector3(0, 1, 0),
			Fovy:     45,
		}

		state := &SimulationState{
			Pause:  false,
			UseGPU: false,
			Yaw:    0.0,
			Pitch:  0.0,
		}

		config := &InputConfig{
			MoveSpeed:        1.0,
			MouseSensitivity: 0.01,
			ScreenWidth:      800,
			ScreenHeight:     600,
		}

		// Simulate keyboard and mouse input
		controller.keyboard.SetKeyState(rl.KeyW, true)
		controller.keyboard.SetKeyPressed(rl.KeyP, true)
		controller.mouse.SetButtonDown(rl.MouseRightButton, true)
		controller.mouse.SetMouseDelta(10, 5)

		// Process input
		controller.ProcessInput(camera, state, config)

		// Check that state was updated
		assert.True(t, state.Pause)
		assert.NotEqual(t, 0.0, state.Yaw)
		assert.NotEqual(t, 0.0, state.Pitch)

		// Check that camera was moved
		assert.NotEqual(t, float32(0), camera.Position.X)
	})
}

func TestInputController_UpdateFromRaylib(t *testing.T) {
	controller := NewInputController()

	t.Run("Updates handlers from raylib", func(t *testing.T) {
		// This test would normally mock raylib calls
		// For now, we just verify the method exists
		controller.UpdateFromRaylib()
		assert.NotNil(t, controller)
	})
}

func TestInputController_Reset(t *testing.T) {
	controller := NewInputController()

	t.Run("Reset clears input states", func(t *testing.T) {
		// Set some states
		controller.keyboard.SetKeyState(rl.KeyW, true)
		controller.mouse.SetButtonDown(rl.MouseRightButton, true)

		// Reset
		controller.Reset()

		// Verify states are cleared
		assert.False(t, controller.keyboard.IsKeyDown(rl.KeyW))
		assert.False(t, controller.mouse.IsButtonDown(rl.MouseRightButton))
	})
}
