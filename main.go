package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

const (
	screenWidth  = 1920
	screenHeight = 1080
	gridSize     = 20
	gridSpacing  = 1.0
	moveSpeed    = 0.1
)

var (
	pause            = false // Press 'P' to toggle
	mouseSensitivity = float32(0.003)
	yaw              = float32(0.0)
	pitch            = float32(0.0)
)

func processInput(camera *rl.Camera3D) {
	// Toggle Pause
	if rl.IsKeyPressed(rl.KeyP) {
		pause = !pause
	}

	// Handle mouse rotation when right button is held
	if rl.IsMouseButtonDown(rl.MouseRightButton) {
		mouseDelta := rl.GetMouseDelta()
		yaw += mouseDelta.X * mouseSensitivity
		pitch -= mouseDelta.Y * mouseSensitivity

		// Clamp pitch to prevent flipping
		if pitch > 1.5 {
			pitch = 1.5
		}
		if pitch < -1.5 {
			pitch = -1.5
		}

		// Update camera target based on yaw and pitch
		camera.Target.X = camera.Position.X + float32(math.Cos(float64(yaw))*math.Cos(float64(pitch)))
		camera.Target.Y = camera.Position.Y + float32(math.Sin(float64(pitch)))
		camera.Target.Z = camera.Position.Z + float32(math.Sin(float64(yaw))*math.Cos(float64(pitch)))
	}

	// Calculate forward and right vectors in XZ plane only
	forward := rl.NewVector3(float32(math.Cos(float64(yaw))), 0, float32(math.Sin(float64(yaw))))
	right := rl.NewVector3(float32(math.Cos(float64(yaw-1.5708))), 0, float32(math.Sin(float64(yaw-1.5708)))) // yaw - π/2

	// WASD movement in XZ plane only
	if rl.IsKeyDown(rl.KeyW) {
		camera.Position.X += forward.X * moveSpeed
		camera.Position.Z += forward.Z * moveSpeed
		camera.Target.X += forward.X * moveSpeed
		camera.Target.Z += forward.Z * moveSpeed
	}
	if rl.IsKeyDown(rl.KeyS) {
		camera.Position.X -= forward.X * moveSpeed
		camera.Position.Z -= forward.Z * moveSpeed
		camera.Target.X -= forward.X * moveSpeed
		camera.Target.Z -= forward.Z * moveSpeed
	}
	if rl.IsKeyDown(rl.KeyA) {
		camera.Position.X += right.X * moveSpeed
		camera.Position.Z += right.Z * moveSpeed
		camera.Target.X += right.X * moveSpeed
		camera.Target.Z += right.Z * moveSpeed
	}
	if rl.IsKeyDown(rl.KeyD) {
		camera.Position.X -= right.X * moveSpeed
		camera.Position.Z -= right.Z * moveSpeed
		camera.Target.X -= right.X * moveSpeed
		camera.Target.Z -= right.Z * moveSpeed
	}

	// Q/E movement in Y axis only
	if rl.IsKeyDown(rl.KeyQ) {
		camera.Position.Y -= moveSpeed
		camera.Target.Y -= moveSpeed
	}
	if rl.IsKeyDown(rl.KeyE) {
		camera.Position.Y += moveSpeed
		camera.Target.Y += moveSpeed
	}
}

func main() {
	// Initialize window
	rl.InitWindow(screenWidth, screenHeight, "3D Grid - Raylib")
	defer rl.CloseWindow()

	// Set up camera
	camera := rl.Camera3D{
		Position:   rl.NewVector3(10.0, 10.0, 10.0),
		Target:     rl.NewVector3(0.0, 0.0, 0.0),
		Up:         rl.NewVector3(0.0, 1.0, 0.0),
		Fovy:       65.0,
		Projection: rl.CameraPerspective,
	}

	rl.HideCursor()
	rl.SetClipPlanes(0.01, 100000000)
	rl.SetTargetFPS(60)

	// Main game loop
	for !rl.WindowShouldClose() {
		// Handle input
		processInput(&camera)

		// Draw
		draw(&camera)
	}
}

func draw(camera *rl.Camera) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)

	rl.BeginMode3D(*camera)

	// Draw 3D grid
	drawGrid3D(gridSize, gridSpacing)

	// Draw coordinate axes
	rl.DrawLine3D(rl.NewVector3(0, 0, 0), rl.NewVector3(5, 0, 0), rl.Red)   // X axis
	rl.DrawLine3D(rl.NewVector3(0, 0, 0), rl.NewVector3(0, 5, 0), rl.Green) // Y axis
	rl.DrawLine3D(rl.NewVector3(0, 0, 0), rl.NewVector3(0, 0, 5), rl.Blue)  // Z axis

	rl.EndMode3D()

	// Draw UI
	rl.DrawText("3D Grid with Raylib", 10, 10, 20, rl.DarkGray)

	rl.EndDrawing()
}

func drawGrid3D(size int, spacing float32) {
	halfSize := float32(size) * spacing / 2.0

	// Draw grid lines on XZ plane (horizontal grid)
	for i := 0; i <= size; i++ {
		pos := float32(i)*spacing - halfSize

		// Lines parallel to X axis
		rl.DrawLine3D(
			rl.NewVector3(-halfSize, 0, pos),
			rl.NewVector3(halfSize, 0, pos),
			rl.Gray,
		)

		// Lines parallel to Z axis
		rl.DrawLine3D(
			rl.NewVector3(pos, 0, -halfSize),
			rl.NewVector3(pos, 0, halfSize),
			rl.Gray,
		)
	}
}
