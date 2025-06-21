package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 1920
	screenHeight = 1080
	gridSize     = 20
	gridSpacing  = 1.0
	moveSpeed    = 0.1
)

var (
	pause = false // Press 'P' to toggle
)

func processInput(camera *rl.Camera3D) {
	// Toggle Pause
	if rl.IsKeyPressed(rl.KeyP) {
		pause = !pause
	}

	// Calculate forward and right vectors in XZ plane
	forward := rl.Vector3Subtract(camera.Target, camera.Position)
	forward.Y = 0 // Keep movement in XZ plane
	forward = rl.Vector3Normalize(forward)
	
	right := rl.Vector3CrossProduct(forward, rl.NewVector3(0, 1, 0))
	right = rl.Vector3Normalize(right)

	// WASD movement in XZ plane
	if rl.IsKeyDown(rl.KeyW) {
		movement := rl.Vector3Scale(forward, moveSpeed)
		camera.Position = rl.Vector3Add(camera.Position, movement)
		camera.Target = rl.Vector3Add(camera.Target, movement)
	}
	if rl.IsKeyDown(rl.KeyS) {
		movement := rl.Vector3Scale(forward, -moveSpeed)
		camera.Position = rl.Vector3Add(camera.Position, movement)
		camera.Target = rl.Vector3Add(camera.Target, movement)
	}
	if rl.IsKeyDown(rl.KeyA) {
		movement := rl.Vector3Scale(right, -moveSpeed)
		camera.Position = rl.Vector3Add(camera.Position, movement)
		camera.Target = rl.Vector3Add(camera.Target, movement)
	}
	if rl.IsKeyDown(rl.KeyD) {
		movement := rl.Vector3Scale(right, moveSpeed)
		camera.Position = rl.Vector3Add(camera.Position, movement)
		camera.Target = rl.Vector3Add(camera.Target, movement)
	}

	// Q/E movement in Y axis (up/down)
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
		if !rl.IsMouseButtonDown(rl.MouseRightButton) {
			rl.SetMousePosition(screenWidth/2, screenHeight/2)
		}

		// Update camera
		rl.UpdateCamera(&camera, rl.CameraCustom)

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
