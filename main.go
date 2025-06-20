package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1200
	screenHeight = 800
	G            = 6.67430e-7 // 중력 상수 (스케일링됨)
	c            = 299792458  // 광속 (스케일링됨)
	gridSize     = 20
	maxParticles = 50
)

// Vector3D represents a 3D vector
type Vector3D struct {
	X, Y, Z float64
}

func (v Vector3D) Add(other Vector3D) Vector3D {
	return Vector3D{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

func (v Vector3D) Sub(other Vector3D) Vector3D {
	return Vector3D{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

func (v Vector3D) Mul(scalar float64) Vector3D {
	return Vector3D{v.X * scalar, v.Y * scalar, v.Z * scalar}
}

func (v Vector3D) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vector3D) Normalize() Vector3D {
	mag := v.Magnitude()
	if mag == 0 {
		return Vector3D{0, 0, 0}
	}
	return Vector3D{v.X / mag, v.Y / mag, v.Z / mag}
}

// Particle represents a massive object in spacetime
type Particle struct {
	Position    Vector3D
	Velocity    Vector3D
	Mass        float64
	Radius      float64
	Color       color.RGBA
	Trail       []Vector3D
	MaxTrailLen int
}

// SpacetimeGrid represents the curvature of spacetime
type SpacetimeGrid struct {
	Width, Height int
	Curvature     [][]float64
}

// Game implements ebiten.Game interface
type Game struct {
	particles     []*Particle
	grid          *SpacetimeGrid
	camera        Vector3D
	zoom          float64
	paused        bool
	showGrid      bool
	timeScale     float64
	selectedIndex int
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	game := &Game{
		particles:     make([]*Particle, 0),
		grid:          NewSpacetimeGrid(screenWidth/gridSize, screenHeight/gridSize),
		camera:        Vector3D{0, 0, 0},
		zoom:          1.0,
		paused:        false,
		showGrid:      true,
		timeScale:     1.0,
		selectedIndex: -1,
	}

	// 초기 파티클들 생성
	game.initializeParticles()

	return game
}

func NewSpacetimeGrid(width, height int) *SpacetimeGrid {
	curvature := make([][]float64, height)
	for i := range curvature {
		curvature[i] = make([]float64, width)
	}

	return &SpacetimeGrid{
		Width:     width,
		Height:    height,
		Curvature: curvature,
	}
}

func (g *Game) initializeParticles() {
	// Central massive body (black hole or star)
	g.particles = append(g.particles, &Particle{
		Position:    Vector3D{screenWidth / 2, screenHeight / 2, 0},
		Velocity:    Vector3D{0, 0, 0},
		Mass:        1e12,
		Radius:      20,
		Color:       color.RGBA{255, 255, 0, 255}, // Yellow
		Trail:       make([]Vector3D, 0),
		MaxTrailLen: 100,
	})

	// Orbiting smaller particles
	for i := 0; i < 8; i++ {
		angle := float64(i) * 2 * math.Pi / 8
		distance := 150.0 + rand.Float64()*200

		pos := Vector3D{
			screenWidth/2 + distance*math.Cos(angle),
			screenHeight/2 + distance*math.Sin(angle),
			rand.Float64()*50 - 25,
		}

		// Calculate circular orbital velocity
		orbitalSpeed := math.Sqrt(G*g.particles[0].Mass/distance) * 0.1
		vel := Vector3D{
			-orbitalSpeed * math.Sin(angle),
			orbitalSpeed * math.Cos(angle),
			rand.Float64()*10 - 5,
		}

		g.particles = append(g.particles, &Particle{
			Position:    pos,
			Velocity:    vel,
			Mass:        1e8 + rand.Float64()*1e9,
			Radius:      5 + rand.Float64()*5,
			Color:       color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255},
			Trail:       make([]Vector3D, 0),
			MaxTrailLen: 50,
		})
	}
}

func (g *Game) Update() error {
	// 입력 처리
	g.handleInput()

	if !g.paused {
		// 물리 시뮬레이션 업데이트
		g.updatePhysics()

		// 시공간 곡률 계산
		g.updateSpacetimeCurvature()

		// 충돌 감지
		g.handleCollisions()
	}

	return nil
}

func (g *Game) handleInput() {
	// Pause/Resume
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.paused = !g.paused
	}

	// Toggle grid display
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.showGrid = !g.showGrid
	}

	// Time scale adjustment
	if ebiten.IsKeyPressed(ebiten.KeyEqual) {
		g.timeScale = math.Min(g.timeScale*1.1, 5.0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) {
		g.timeScale = math.Max(g.timeScale*0.9, 0.1)
	}

	// Zoom adjustment
	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		g.zoom = math.Min(g.zoom*1.05, 3.0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		g.zoom = math.Max(g.zoom*0.95, 0.3)
	}

	// Camera movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.camera.X -= 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.camera.X += 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.camera.Y -= 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.camera.Y += 5
	}

	// Add new particle (mouse click)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && len(g.particles) < maxParticles {
		x, y := ebiten.CursorPosition()
		g.addParticle(float64(x), float64(y))
	}

	// Select particle (right click)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		x, y := ebiten.CursorPosition()
		g.selectParticle(float64(x), float64(y))
	}
}

func (g *Game) addParticle(x, y float64) {
	// Convert screen coordinates to world coordinates
	worldX := (x-screenWidth/2)/g.zoom + g.camera.X + screenWidth/2
	worldY := (y-screenHeight/2)/g.zoom + g.camera.Y + screenHeight/2

	particle := &Particle{
		Position:    Vector3D{worldX, worldY, rand.Float64()*100 - 50},
		Velocity:    Vector3D{rand.Float64()*20 - 10, rand.Float64()*20 - 10, rand.Float64()*10 - 5},
		Mass:        1e8 + rand.Float64()*5e8,
		Radius:      3 + rand.Float64()*7,
		Color:       color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255},
		Trail:       make([]Vector3D, 0),
		MaxTrailLen: 30,
	}

	g.particles = append(g.particles, particle)
}

func (g *Game) selectParticle(x, y float64) {
	worldX := (x-screenWidth/2)/g.zoom + g.camera.X + screenWidth/2
	worldY := (y-screenHeight/2)/g.zoom + g.camera.Y + screenHeight/2

	minDist := math.Inf(1)
	selectedIdx := -1

	for i, p := range g.particles {
		dist := math.Sqrt(math.Pow(p.Position.X-worldX, 2) + math.Pow(p.Position.Y-worldY, 2))
		if dist < p.Radius*2 && dist < minDist {
			minDist = dist
			selectedIdx = i
		}
	}

	g.selectedIndex = selectedIdx
}

func (g *Game) updatePhysics() {
	dt := g.timeScale / 60.0 // Based on 60 FPS

	// Calculate gravity for each particle
	for i, p1 := range g.particles {
		force := Vector3D{0, 0, 0}

		for j, p2 := range g.particles {
			if i == j {
				continue
			}

			// Calculate distance vector
			r := p2.Position.Sub(p1.Position)
			distance := r.Magnitude()

			if distance < p1.Radius+p2.Radius {
				continue // Collision handling is separate
			}

			// Newtonian gravity + General Relativity correction
			gravityMagnitude := G * p1.Mass * p2.Mass / (distance * distance)

			// General Relativity correction (Schwarzschild radius approximation)
			schwarzschildRadius := 2 * G * p2.Mass / (c * c)
			relativisticCorrection := 1.0 + 3*schwarzschildRadius/(2*distance)

			gravityMagnitude *= relativisticCorrection

			// Direction vector
			direction := r.Normalize()
			gravityForce := direction.Mul(gravityMagnitude)

			force = force.Add(gravityForce)
		}

		// Acceleration = Force / Mass
		acceleration := force.Mul(1.0 / p1.Mass)

		// Update velocity (Verlet integration)
		p1.Velocity = p1.Velocity.Add(acceleration.Mul(dt))

		// Update position
		p1.Position = p1.Position.Add(p1.Velocity.Mul(dt))

		// Add to trail
		p1.Trail = append(p1.Trail, p1.Position)
		if len(p1.Trail) > p1.MaxTrailLen {
			p1.Trail = p1.Trail[1:]
		}
	}
}

func (g *Game) updateSpacetimeCurvature() {
	// Initialize grid
	for i := range g.grid.Curvature {
		for j := range g.grid.Curvature[i] {
			g.grid.Curvature[i][j] = 0
		}
	}

	// Calculate spacetime curvature based on each particle's mass
	for _, p := range g.particles {
		// Calculate influence on surrounding grid points
		for gy := 0; gy < g.grid.Height; gy++ {
			for gx := 0; gx < g.grid.Width; gx++ {
				// Grid point position in world coordinates
				pointX := float64(gx * gridSize)
				pointY := float64(gy * gridSize)

				// Distance from particle to grid point
				dx := p.Position.X - pointX
				dy := p.Position.Y - pointY
				distance := math.Sqrt(dx*dx + dy*dy)

				if distance > 0 {
					// Spacetime curvature (proportional to mass, inversely proportional to distance squared)
					curvature := p.Mass / (distance*distance + 1000) // Prevent singularity
					g.grid.Curvature[gy][gx] += curvature * 1e-10    // Scaling
				}
			}
		}
	}
}

func (g *Game) handleCollisions() {
	for i := 0; i < len(g.particles); i++ {
		for j := i + 1; j < len(g.particles); j++ {
			p1, p2 := g.particles[i], g.particles[j]

			distance := p1.Position.Sub(p2.Position).Magnitude()

			if distance < p1.Radius+p2.Radius {
				// Collision occurred - perfectly inelastic collision
				totalMass := p1.Mass + p2.Mass

				// Calculate center of mass
				newPos := p1.Position.Mul(p1.Mass).Add(p2.Position.Mul(p2.Mass)).Mul(1.0 / totalMass)
				newVel := p1.Velocity.Mul(p1.Mass).Add(p2.Velocity.Mul(p2.Mass)).Mul(1.0 / totalMass)

				// Create new particle
				newParticle := &Particle{
					Position:    newPos,
					Velocity:    newVel,
					Mass:        totalMass,
					Radius:      math.Pow(math.Pow(p1.Radius, 3)+math.Pow(p2.Radius, 3), 1.0/3.0),
					Color:       color.RGBA{255, 255, 255, 255}, // Display in white
					Trail:       make([]Vector3D, 0),
					MaxTrailLen: int(math.Max(float64(p1.MaxTrailLen), float64(p2.MaxTrailLen))),
				}

				// Remove existing particles and add new particle
				g.particles = append(g.particles[:i], g.particles[i+1:]...)
				if j > i {
					j--
				}
				g.particles = append(g.particles[:j], g.particles[j+1:]...)
				g.particles = append(g.particles, newParticle)

				// Adjust selected index
				if g.selectedIndex == i || g.selectedIndex == j {
					g.selectedIndex = len(g.particles) - 1
				} else if g.selectedIndex > j {
					g.selectedIndex -= 2
				} else if g.selectedIndex > i {
					g.selectedIndex--
				}

				return // Process only one collision at a time
			}
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 20, 255}) // Dark blue background

	// Draw spacetime grid
	if g.showGrid {
		g.drawSpacetimeGrid(screen)
	}

	// Draw particle trails
	g.drawTrails(screen)

	// Draw particles
	g.drawParticles(screen)

	// Display UI information
	g.drawUI(screen)
}

func (g *Game) drawSpacetimeGrid(screen *ebiten.Image) {
	// Draw grid lines first (base spacetime fabric)
	for i := 0; i <= g.grid.Height; i++ {
		for j := 0; j <= g.grid.Width; j++ {
			worldX := float64(j * gridSize)
			worldY := float64(i * gridSize)

			screenX := (worldX-g.camera.X)*g.zoom + screenWidth/2
			screenY := (worldY-g.camera.Y)*g.zoom + screenHeight/2

			// Skip if outside screen bounds
			if screenX < -50 || screenX > screenWidth+50 || screenY < -50 || screenY > screenHeight+50 {
				continue
			}

			// Draw horizontal grid lines
			if i < g.grid.Height && j < g.grid.Width {
				nextWorldX := float64((j + 1) * gridSize)
				nextScreenX := (nextWorldX-g.camera.X)*g.zoom + screenWidth/2

				// Get curvature for height displacement
				curvature := g.grid.Curvature[i][j]
				nextCurvature := float64(0)
				if j+1 < g.grid.Width {
					nextCurvature = g.grid.Curvature[i][j+1]
				}

				height := curvature * 5e10
				nextHeight := nextCurvature * 5e10

				vector.StrokeLine(screen, 
					float32(screenX), float32(screenY-height), 
					float32(nextScreenX), float32(screenY-nextHeight), 
					1, color.RGBA{50, 50, 100, 100}, false)
			}

			// Draw vertical grid lines
			if i < g.grid.Height && j < g.grid.Width {
				nextWorldY := float64((i + 1) * gridSize)
				nextScreenY := (nextWorldY-g.camera.Y)*g.zoom + screenHeight/2

				// Get curvature for height displacement
				curvature := g.grid.Curvature[i][j]
				nextCurvature := float64(0)
				if i+1 < g.grid.Height {
					nextCurvature = g.grid.Curvature[i+1][j]
				}

				height := curvature * 5e10
				nextHeight := nextCurvature * 5e10

				vector.StrokeLine(screen, 
					float32(screenX), float32(screenY-height), 
					float32(screenX), float32(nextScreenY-nextHeight), 
					1, color.RGBA{50, 50, 100, 100}, false)
			}
		}
	}

	// Draw curvature intensity points
	for i := 0; i < g.grid.Height; i++ {
		for j := 0; j < g.grid.Width; j++ {
			curvature := g.grid.Curvature[i][j]

			if curvature > 1e-12 {
				worldX := float64(j * gridSize)
				worldY := float64(i * gridSize)

				screenX := (worldX-g.camera.X)*g.zoom + screenWidth/2
				screenY := (worldY-g.camera.Y)*g.zoom + screenHeight/2

				// Skip if outside screen bounds
				if screenX < -50 || screenX > screenWidth+50 || screenY < -50 || screenY > screenHeight+50 {
					continue
				}

				// Color intensity based on curvature
				intensity := math.Min(curvature*1e12, 1.0)
				alpha := uint8(intensity * 200)

				if alpha > 10 {
					// 3D height effect
					height := curvature * 5e10

					// Draw curvature point
					radius := float32(2 + intensity*3)
					vector.DrawFilledCircle(screen, float32(screenX), float32(screenY-height), radius, 
						color.RGBA{255, uint8(100 + intensity*155), 0, alpha}, false)
				}
			}
		}
	}
}

func (g *Game) drawTrails(screen *ebiten.Image) {
	for _, p := range g.particles {
		if len(p.Trail) < 2 {
			continue
		}

		for i := 1; i < len(p.Trail); i++ {
			pos1 := p.Trail[i-1]
			pos2 := p.Trail[i]

			// Convert world coordinates to screen coordinates
			screen1X := (pos1.X-g.camera.X)*g.zoom + screenWidth/2
			screen1Y := (pos1.Y-g.camera.Y)*g.zoom + screenHeight/2
			screen2X := (pos2.X-g.camera.X)*g.zoom + screenWidth/2
			screen2Y := (pos2.Y-g.camera.Y)*g.zoom + screenHeight/2

			// Trail color (gradually fading)
			alpha := uint8(float64(i) / float64(len(p.Trail)) * 100)
			trailColor := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}

			vector.StrokeLine(screen, float32(screen1X), float32(screen1Y), float32(screen2X), float32(screen2Y), 1, trailColor, false)
		}
	}
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for i, p := range g.particles {
		// Convert world coordinates to screen coordinates
		screenX := (p.Position.X-g.camera.X)*g.zoom + screenWidth/2
		screenY := (p.Position.Y-g.camera.Y)*g.zoom + screenHeight/2

		// 3D effect based on Z coordinate
		zOffset := p.Position.Z * 0.1
		screenY -= zOffset

		radius := float32(p.Radius * g.zoom)

		// Highlight selected particle
		if i == g.selectedIndex {
			vector.StrokeCircle(screen, float32(screenX), float32(screenY), radius+3, 2, color.RGBA{255, 255, 255, 255}, false)
		}

		// Draw particle
		vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), radius, p.Color, false)

		// Halo effect for massive particles
		if p.Mass > 5e11 {
			for r := radius + 5; r < radius+15; r += 2 {
				alpha := uint8(50 * (radius + 15 - r) / 10)
				haloColor := color.RGBA{255, 255, 0, alpha}
				vector.StrokeCircle(screen, float32(screenX), float32(screenY), r, 1, haloColor, false)
			}
		}
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Status information
	status := fmt.Sprintf("Particles: %d | Time Scale: %.1fx | Zoom: %.1fx", len(g.particles), g.timeScale, g.zoom)
	if g.paused {
		status += " | PAUSED"
	}
	ebitenutil.DebugPrint(screen, status)

	// Controls
	instructions := []string{
		"Controls:",
		"SPACE: Pause/Resume",
		"G: Toggle Grid Display",
		"+/-: Adjust Time Scale",
		"Z/X: Zoom In/Out",
		"Arrows: Move Camera",
		"Left Click: Add Particle",
		"Right Click: Select Particle",
	}

	for i, instruction := range instructions {
		ebitenutil.DebugPrintAt(screen, instruction, 10, 30+i*15)
	}

	// Selected particle information
	if g.selectedIndex >= 0 && g.selectedIndex < len(g.particles) {
		p := g.particles[g.selectedIndex]
		info := fmt.Sprintf("Selected Particle:\nMass: %.2e\nVelocity: %.1f\nPosition: (%.1f, %.1f, %.1f)",
			p.Mass, p.Velocity.Magnitude(), p.Position.X, p.Position.Y, p.Position.Z)
		ebitenutil.DebugPrintAt(screen, info, screenWidth-200, 30)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("General Relativity Simulation - N-Body Gravity System")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
