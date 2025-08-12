package input

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/stretchr/testify/assert"
)

func TestMouseHandler_ProcessRotation(t *testing.T) {
	handler := NewMouseHandler()

	// Test mouse rotation when right button is held
	t.Run("Right button enables rotation", func(t *testing.T) {
		// Initially, rotation should be disabled
		rotation := handler.ProcessRotation(0.0, 0.0, 0.01)
		assert.False(t, rotation.Active)

		// Enable right button
		handler.SetButtonDown(rl.MouseRightButton, true)
		handler.SetMouseDelta(10, 5)

		rotation = handler.ProcessRotation(0.0, 0.0, 0.01)
		assert.True(t, rotation.Active)
		assert.NotEqual(t, 0.0, rotation.YawDelta)
		assert.NotEqual(t, 0.0, rotation.PitchDelta)
	})

	// Test yaw calculation
	t.Run("Yaw increases with positive X delta", func(t *testing.T) {
		handler := NewMouseHandler()
		handler.SetButtonDown(rl.MouseRightButton, true)
		handler.SetMouseDelta(10, 0)

		rotation := handler.ProcessRotation(0.0, 0.0, 0.01)
		assert.Greater(t, rotation.YawDelta, float32(0.0))
		assert.Equal(t, float32(0.0), rotation.PitchDelta)
	})

	// Test pitch calculation
	t.Run("Pitch decreases with positive Y delta", func(t *testing.T) {
		handler := NewMouseHandler()
		handler.SetButtonDown(rl.MouseRightButton, true)
		handler.SetMouseDelta(0, 10)

		rotation := handler.ProcessRotation(0.0, 0.0, 0.01)
		assert.Equal(t, float32(0.0), rotation.YawDelta)
		assert.Less(t, rotation.PitchDelta, float32(0.0))
	})

	// Test pitch clamping
	t.Run("Pitch is clamped to prevent flipping", func(t *testing.T) {
		handler := NewMouseHandler()
		handler.SetButtonDown(rl.MouseRightButton, true)
		handler.SetMouseDelta(0, 1000) // Large delta

		// Test upper clamp
		rotation := handler.ProcessRotation(0.0, 1.4, 0.01)
		newPitch := 1.4 + rotation.PitchDelta
		assert.LessOrEqual(t, newPitch, float32(1.5))

		// Test lower clamp
		handler.SetMouseDelta(0, -1000)
		rotation = handler.ProcessRotation(0.0, -1.4, 0.01)
		newPitch = -1.4 + rotation.PitchDelta
		assert.GreaterOrEqual(t, newPitch, float32(-1.5))
	})

	// Test mouse centering when not rotating
	t.Run("Mouse should center when not rotating", func(t *testing.T) {
		handler := NewMouseHandler()
		handler.SetButtonDown(rl.MouseRightButton, false)

		rotation := handler.ProcessRotation(0.0, 0.0, 0.01)
		assert.False(t, rotation.Active)
		assert.True(t, rotation.ShouldCenter)
	})
}

func TestMouseHandler_UpdateCamera(t *testing.T) {
	handler := NewMouseHandler()

	t.Run("Updates camera target based on yaw and pitch", func(t *testing.T) {
		camera := &rl.Camera3D{
			Position: rl.NewVector3(0, 0, 0),
			Target:   rl.NewVector3(1, 0, 0),
			Up:       rl.NewVector3(0, 1, 0),
			Fovy:     45,
		}

		handler.UpdateCameraTarget(camera, 0.0, 0.0)

		// Target should be updated based on yaw and pitch
		assert.NotEqual(t, float32(0), camera.Target.X)
	})
}

func TestMouseHandler_Sensitivity(t *testing.T) {
	handler := NewMouseHandler()

	t.Run("Sensitivity affects rotation speed", func(t *testing.T) {
		handler.SetButtonDown(rl.MouseRightButton, true)
		handler.SetMouseDelta(10, 10)

		// Low sensitivity
		rotation1 := handler.ProcessRotation(0.0, 0.0, 0.001)

		// High sensitivity
		rotation2 := handler.ProcessRotation(0.0, 0.0, 0.01)

		assert.Less(t, rotation1.YawDelta, rotation2.YawDelta)
		assert.Less(t, rotation2.PitchDelta, rotation1.PitchDelta) // More negative
	})
}
