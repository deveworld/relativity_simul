package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mjibson/go-dsp/fft"
	"math"
	"math/rand"
)

const (
	screenWidth  = 1920
	screenHeight = 1080

	simulationWidth = 128
	simulationDepth = 128

	numParticles = 5
	gridVisScale = 0.1
	gConstant    = 1.0

	timeStep  = 0.1
	moveSpeed = 0.2
)

var (
	pause            = false // Press 'P' to toggle
	mouseSensitivity = float32(0.003)
	yaw              = float32(3.92699) // Start facing -Z direction
	pitch            = float32(-0.628)  // Start looking slightly down
)

// Particle represents a massive body in the simulation
type Particle struct {
	Position rl.Vector3
	Velocity rl.Vector3
	Mass     float32
	Radius   float32
}

// Simulation holds the entire state of the GR simulation
type Simulation struct {
	Particles       []*Particle
	PotentialGrid   [][]float64 // Stores the scalar potential Φ (proportional to h_00)
	MassDensityGrid [][]float64 // Stores the mass density ρ
	AccelFieldX     [][]float64 // Stores the X component of the acceleration field
	AccelFieldZ     [][]float64 // Stores the Z component of the acceleration field
}

// NewSimulation creates and initializes a new simulation instance
func NewSimulation() *Simulation {
	sim := &Simulation{
		Particles:       make([]*Particle, numParticles),
		PotentialGrid:   make([][]float64, simulationWidth),
		MassDensityGrid: make([][]float64, simulationWidth),
		AccelFieldX:     make([][]float64, simulationWidth),
		AccelFieldZ:     make([][]float64, simulationWidth),
	}

	for i := range sim.PotentialGrid {
		sim.PotentialGrid[i] = make([]float64, simulationDepth)
		sim.MassDensityGrid[i] = make([]float64, simulationDepth)
		sim.AccelFieldX[i] = make([]float64, simulationDepth)
		sim.AccelFieldZ[i] = make([]float64, simulationDepth)
	}

	// Initialize particles with random positions and some mass
	for i := 0; i < numParticles; i++ {
		mass := 20.0 + rand.Float32()*30.0
		sim.Particles[i] = &Particle{
			Position: rl.NewVector3(
				(rand.Float32()-0.5)*simulationWidth*0.8,
				0,
				(rand.Float32()-0.5)*simulationDepth*0.8,
			),
			Velocity: rl.NewVector3(0, 0, 0),
			Mass:     mass,
			Radius:   float32(math.Pow(float64(mass/20.0), 1.0/3.0)) * 0.5, // Radius scales with mass
		}
	}

	// Add a large central mass
	// sim.Particles[0].Mass = 1000
	// sim.Particles[0].Radius = float32(math.Pow(float64(sim.Particles[0].Mass/20.0), 1.0/3.0)) * 0.5
	// sim.Particles[0].Position = rl.NewVector3(0,0,0)

	return sim
}

// Update runs one full step of the simulation
func (s *Simulation) Update() {
	// 1. Kick (Leapfrog integrator part 1)
	s.updateVelocities(0.5 * timeStep)

	// 2. Drift (Leapfrog integrator)
	s.updatePositions(timeStep)

	// 3. Calculate new accelerations for the next step
	s.calculateAccelerationField()

	// 4. Kick (Leapfrog integrator part 2)
	s.updateVelocities(0.5 * timeStep)
}

// calculateAccelerationField performs the main PM method steps
func (s *Simulation) calculateAccelerationField() {
	// Step 1: Deposit mass onto the grid (Cloud-in-Cell)
	s.depositMass()
	// Step 2: Solve for potential Φ using FFT
	s.solvePotential()
	//s.solvePotentialGPU()
	// Step 3: Calculate acceleration (a = -∇Φ) from the potential field
	s.calculateGradient()
}

// depositMass distributes particle mass to the grid using Cloud-in-Cell
func (s *Simulation) depositMass() {
	// Clear the grid first
	for i := range s.MassDensityGrid {
		for j := range s.MassDensityGrid[i] {
			s.MassDensityGrid[i][j] = 0
		}
	}

	for _, p := range s.Particles {
		// Find grid cell coordinates and fractional parts
		gx := float64(p.Position.X) + float64(simulationWidth)/2.0
		gz := float64(p.Position.Z) + float64(simulationDepth)/2.0
		i := int(gx)
		j := int(gz)
		fx := gx - float64(i)
		fz := gz - float64(j)

		// Distribute mass to 4 nearest cells (Cloud-in-Cell)
		if i >= 0 && i < simulationWidth-1 && j >= 0 && j < simulationDepth-1 {
			s.MassDensityGrid[i][j] += float64(p.Mass) * (1 - fx) * (1 - fz)
			s.MassDensityGrid[i+1][j] += float64(p.Mass) * fx * (1 - fz)
			s.MassDensityGrid[i][j+1] += float64(p.Mass) * (1 - fx) * fz
			s.MassDensityGrid[i+1][j+1] += float64(p.Mass) * fx * fz
		}
	}
}

// solvePotential solves ∇²Φ = 2πGρ using FFT
func (s *Simulation) solvePotential() {
	// Convert mass density grid to complex numbers for FFT
	complexGrid := make([][]complex128, simulationWidth)
	for i := range complexGrid {
		complexGrid[i] = make([]complex128, simulationDepth)
		for j := range complexGrid[i] {
			complexGrid[i][j] = complex(s.MassDensityGrid[i][j], 0)
		}
	}

	// 2D FFT of the mass density
	fftGrid := fft.FFT2(complexGrid)

	// Solve in Fourier space: Φ̂(k) = -2πG * ρ̂(k) / |k|²
	// Note: We use a different sign convention where a = -∇Φ, so we solve for ∇²Φ = -2πGρ
	kxFactor := 2.0 * math.Pi / float64(simulationWidth)
	kzFactor := 2.0 * math.Pi / float64(simulationDepth)

	for u := 0; u < simulationWidth; u++ {
		for v := 0; v < simulationDepth; v++ {
			// Calculate wave vector k
			kx := float64(u)
			if u > simulationWidth/2 {
				kx = float64(u - simulationWidth)
			}
			kz := float64(v)
			if v > simulationDepth/2 {
				kz = float64(v - simulationDepth)
			}

			kSquared := (kx*kxFactor)*(kx*kxFactor) + (kz*kzFactor)*(kz*kzFactor)

			if kSquared == 0 {
				fftGrid[u][v] = 0 // Ignore the DC component (average potential)
			} else {
				// The division by k² performs the integration of the Poisson equation
				scalingFactor := -2.0 * math.Pi * gConstant / kSquared
				fftGrid[u][v] *= complex(scalingFactor, 0)
			}
		}
	}

	// Inverse 2D FFT to get the potential grid in real space
	potentialComplex := fft.IFFT2(fftGrid)

	// Copy real part to our potential grid
	for i := range s.PotentialGrid {
		for j := range s.PotentialGrid[i] {
			s.PotentialGrid[i][j] = real(potentialComplex[i][j])
		}
	}
}

func (s *Simulation) solvePotentialGPU() {
}

// calculateGradient computes acceleration a = -∇Φ using central differences
func (s *Simulation) calculateGradient() {
	for i := 0; i < simulationWidth; i++ {
		for j := 0; j < simulationDepth; j++ {
			// Use modulo arithmetic for periodic (wrapping) boundaries
			prevI := (i - 1 + simulationWidth) % simulationWidth
			nextI := (i + 1) % simulationWidth
			prevJ := (j - 1 + simulationDepth) % simulationDepth
			nextJ := (j + 1) % simulationDepth

			// Central difference for gradient with periodic boundaries
			s.AccelFieldX[i][j] = -(s.PotentialGrid[nextI][j] - s.PotentialGrid[prevI][j]) / 2.0
			s.AccelFieldZ[i][j] = -(s.PotentialGrid[i][nextJ] - s.PotentialGrid[i][prevJ]) / 2.0
		}
	}
}

// updatePositions updates the positions of all particles (Drift step)
func (s *Simulation) updatePositions(dt float32) {
	for _, p := range s.Particles {
		p.Position.X += p.Velocity.X * dt
		p.Position.Z += p.Velocity.Z * dt
		// Boundary conditions - wrap around
		if p.Position.X > float32(simulationWidth)/2.0 {
			p.Position.X = -float32(simulationWidth) / 2.0
		}
		if p.Position.X < -float32(simulationWidth)/2.0 {
			p.Position.X = float32(simulationWidth) / 2.0
		}
		if p.Position.Z > float32(simulationDepth)/2.0 {
			p.Position.Z = -float32(simulationDepth) / 2.0
		}
		if p.Position.Z < -float32(simulationDepth)/2.0 {
			p.Position.Z = float32(simulationDepth) / 2.0
		}
	}
}

// updateVelocities updates particle velocities based on acceleration field (Kick step)
func (s *Simulation) updateVelocities(dt float32) {
	for _, p := range s.Particles {
		// Find grid cell coordinates and fractional parts for interpolation
		gx := float64(p.Position.X) + float64(simulationWidth)/2.0
		gz := float64(p.Position.Z) + float64(simulationDepth)/2.0
		i := int(gx)
		j := int(gz)
		fx := gx - float64(i)
		fz := gz - float64(j)

		var ax, az float64
		// Interpolate acceleration from the grid to the particle's position (trilinear)
		if i >= 0 && i < simulationWidth-1 && j >= 0 && j < simulationDepth-1 {
			ax1 := s.AccelFieldX[i][j]*(1-fz) + s.AccelFieldX[i][j+1]*fz
			ax2 := s.AccelFieldX[i+1][j]*(1-fz) + s.AccelFieldX[i+1][j+1]*fz
			ax = ax1*(1-fx) + ax2*fx

			az1 := s.AccelFieldZ[i][j]*(1-fz) + s.AccelFieldZ[i][j+1]*fz
			az2 := s.AccelFieldZ[i+1][j]*(1-fz) + s.AccelFieldZ[i+1][j+1]*fz
			az = az1*(1-fx) + az2*fx
		}

		p.Velocity.X += float32(ax) * dt
		p.Velocity.Z += float32(az) * dt
	}
}

func processInput(camera *rl.Camera3D) {
	// Toggle Pause
	if rl.IsKeyPressed(rl.KeyP) {
		pause = !pause
	}
	// Handle mouse rotation when right button is held
	if !rl.IsMouseButtonDown(rl.MouseRightButton) {
		rl.SetMousePosition(screenWidth/2, screenHeight/2)
	} else {
		delta := rl.GetMouseDelta()
		yaw += delta.X * mouseSensitivity
		pitch -= delta.Y * mouseSensitivity
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
	rl.InitWindow(screenWidth, screenHeight, "Golang GR Simulation - (2+1)D Spacetime")
	defer rl.CloseWindow()

	// Set up camera
	camera := rl.Camera3D{
		Position:   rl.NewVector3(50.0, 50.0, 50.0),
		Target:     rl.NewVector3(0.0, 0.0, 0.0),
		Up:         rl.NewVector3(0.0, 1.0, 0.0),
		Fovy:       65.0,
		Projection: rl.CameraPerspective,
	}

	// Create the simulation
	simulation := NewSimulation()

	rl.HideCursor()
	rl.SetClipPlanes(0.1, 10000.0)
	rl.SetTargetFPS(60)
	// Main game loop
	for !rl.WindowShouldClose() {
		// Handle input
		processInput(&camera)

		// Update simulation state if not paused
		if !pause {
			simulation.Update()
		}
		// Draw the scene
		draw(&camera, simulation)
	}
}

func draw(camera *rl.Camera, sim *Simulation) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)

	rl.BeginMode3D(*camera)

	// Draw the deformed spacetime grid
	drawDeformedGrid(sim)

	// Draw the particles
	for _, p := range sim.Particles {
		rl.DrawSphere(p.Position, p.Radius, rl.Gold)
	}

	// Draw coordinate axes
	rl.DrawLine3D(rl.NewVector3(0, 0, 0), rl.NewVector3(5, 0, 0), rl.Red)   // X axis
	rl.DrawLine3D(rl.NewVector3(0, 0, 0), rl.NewVector3(0, 5, 0), rl.Green) // Y axis
	rl.DrawLine3D(rl.NewVector3(0, 0, 0), rl.NewVector3(0, 0, 5), rl.Blue)  // Z axis

	rl.EndMode3D()

	// Draw UI
	rl.DrawText("GR (Weak-Field) N-Body Simulation", 10, 10, 20, rl.Lime)
	rl.DrawText(fmt.Sprintf("Particles: %d", numParticles), 10, 40, 20, rl.White)
	rl.DrawText("Right-click + Mouse to look", 10, 70, 20, rl.White)
	rl.DrawText("W,A,S,D,Q,E to move", 10, 100, 20, rl.White)
	if pause {
		rl.DrawText("PAUSED (Press P to unpause)", screenWidth/2-150, screenHeight/2-10, 20, rl.Yellow)
	}
	rl.DrawFPS(screenWidth-100, 10)

	rl.EndDrawing()
}

func drawDeformedGrid(sim *Simulation) {
	gridColor := rl.NewColor(50, 50, 100, 255)

	// Draw lines parallel to Z axis
	for i := 0; i < simulationWidth; i++ {
		for j := 0; j < simulationDepth-1; j++ {
			p1X := float32(i) - float32(simulationWidth)/2.0
			p1Z := float32(j) - float32(simulationDepth)/2.0
			p1Y := float32(sim.PotentialGrid[i][j] * gridVisScale)

			p2X := float32(i) - float32(simulationWidth)/2.0
			p2Z := float32(j+1) - float32(simulationDepth)/2.0
			p2Y := float32(sim.PotentialGrid[i][j+1] * gridVisScale)

			rl.DrawLine3D(rl.NewVector3(p1X, p1Y, p1Z), rl.NewVector3(p2X, p2Y, p2Z), gridColor)
		}
	}

	// Draw lines parallel to X axis
	for j := 0; j < simulationDepth; j++ {
		for i := 0; i < simulationWidth-1; i++ {
			p1X := float32(i) - float32(simulationWidth)/2.0
			p1Z := float32(j) - float32(simulationDepth)/2.0
			p1Y := float32(sim.PotentialGrid[i][j] * gridVisScale)

			p2X := float32(i+1) - float32(simulationWidth)/2.0
			p2Z := float32(j) - float32(simulationDepth)/2.0
			p2Y := float32(sim.PotentialGrid[i+1][j] * gridVisScale)

			rl.DrawLine3D(rl.NewVector3(p1X, p1Y, p1Z), rl.NewVector3(p2X, p2Y, p2Z), gridColor)
		}
	}
}
