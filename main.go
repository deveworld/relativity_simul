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
	// 중앙에 큰 질량체 (블랙홀 또는 별)
	g.particles = append(g.particles, &Particle{
		Position:    Vector3D{screenWidth / 2, screenHeight / 2, 0},
		Velocity:    Vector3D{0, 0, 0},
		Mass:        1e12,
		Radius:      20,
		Color:       color.RGBA{255, 255, 0, 255}, // 노란색
		Trail:       make([]Vector3D, 0),
		MaxTrailLen: 100,
	})

	// 궤도를 도는 작은 파티클들
	for i := 0; i < 8; i++ {
		angle := float64(i) * 2 * math.Pi / 8
		distance := 150.0 + rand.Float64()*200

		pos := Vector3D{
			screenWidth/2 + distance*math.Cos(angle),
			screenHeight/2 + distance*math.Sin(angle),
			rand.Float64()*50 - 25,
		}

		// 원형 궤도 속도 계산
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
	// 일시정지/재생
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.paused = !g.paused
	}

	// 그리드 표시 토글
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.showGrid = !g.showGrid
	}

	// 시간 스케일 조정
	if ebiten.IsKeyPressed(ebiten.KeyEqual) {
		g.timeScale = math.Min(g.timeScale*1.1, 5.0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) {
		g.timeScale = math.Max(g.timeScale*0.9, 0.1)
	}

	// 줌 조정
	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		g.zoom = math.Min(g.zoom*1.05, 3.0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		g.zoom = math.Max(g.zoom*0.95, 0.3)
	}

	// 카메라 이동
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

	// 새 파티클 추가 (마우스 클릭)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && len(g.particles) < maxParticles {
		x, y := ebiten.CursorPosition()
		g.addParticle(float64(x), float64(y))
	}

	// 파티클 선택 (우클릭)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		x, y := ebiten.CursorPosition()
		g.selectParticle(float64(x), float64(y))
	}
}

func (g *Game) addParticle(x, y float64) {
	// 화면 좌표를 월드 좌표로 변환
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
	dt := g.timeScale / 60.0 // 60 FPS 기준

	// 각 파티클에 대해 중력 계산
	for i, p1 := range g.particles {
		force := Vector3D{0, 0, 0}

		for j, p2 := range g.particles {
			if i == j {
				continue
			}

			// 거리 벡터 계산
			r := p2.Position.Sub(p1.Position)
			distance := r.Magnitude()

			if distance < p1.Radius+p2.Radius {
				continue // 충돌 처리는 별도로
			}

			// 뉴턴 중력 + 일반상대성이론 보정
			gravityMagnitude := G * p1.Mass * p2.Mass / (distance * distance)

			// 일반상대성이론 보정 (슈바르츠실트 반지름 근사)
			schwarzschildRadius := 2 * G * p2.Mass / (c * c)
			relativisticCorrection := 1.0 + 3*schwarzschildRadius/(2*distance)

			gravityMagnitude *= relativisticCorrection

			// 방향 벡터
			direction := r.Normalize()
			gravityForce := direction.Mul(gravityMagnitude)

			force = force.Add(gravityForce)
		}

		// 가속도 = 힘 / 질량
		acceleration := force.Mul(1.0 / p1.Mass)

		// 속도 업데이트 (Verlet 적분)
		p1.Velocity = p1.Velocity.Add(acceleration.Mul(dt))

		// 위치 업데이트
		p1.Position = p1.Position.Add(p1.Velocity.Mul(dt))

		// 궤적 추가
		p1.Trail = append(p1.Trail, p1.Position)
		if len(p1.Trail) > p1.MaxTrailLen {
			p1.Trail = p1.Trail[1:]
		}
	}
}

func (g *Game) updateSpacetimeCurvature() {
	// 그리드 초기화
	for i := range g.grid.Curvature {
		for j := range g.grid.Curvature[i] {
			g.grid.Curvature[i][j] = 0
		}
	}

	// 각 파티클의 질량에 따른 시공간 곡률 계산
	for _, p := range g.particles {
		gridX := int(p.Position.X / gridSize)
		gridY := int(p.Position.Y / gridSize)

		// 주변 그리드 포인트에 영향 계산
		for dy := -5; dy <= 5; dy++ {
			for dx := -5; dx <= 5; dx++ {
				gx := gridX + dx
				gy := gridY + dy

				if gx >= 0 && gx < g.grid.Width && gy >= 0 && gy < g.grid.Height {
					// 그리드 포인트까지의 거리
					pointX := float64(gx * gridSize)
					pointY := float64(gy * gridSize)

					distance := math.Sqrt(math.Pow(p.Position.X-pointX, 2) + math.Pow(p.Position.Y-pointY, 2))

					if distance > 0 {
						// 시공간 곡률 (질량에 비례, 거리의 제곱에 반비례)
						curvature := p.Mass / (distance*distance + 1000) // 특이점 방지
						g.grid.Curvature[gy][gx] += curvature * 1e-10    // 스케일링
					}
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
				// 충돌 발생 - 완전 비탄성 충돌
				totalMass := p1.Mass + p2.Mass

				// 질량 중심 계산
				newPos := p1.Position.Mul(p1.Mass).Add(p2.Position.Mul(p2.Mass)).Mul(1.0 / totalMass)
				newVel := p1.Velocity.Mul(p1.Mass).Add(p2.Velocity.Mul(p2.Mass)).Mul(1.0 / totalMass)

				// 새로운 파티클 생성
				newParticle := &Particle{
					Position:    newPos,
					Velocity:    newVel,
					Mass:        totalMass,
					Radius:      math.Pow(math.Pow(p1.Radius, 3)+math.Pow(p2.Radius, 3), 1.0/3.0),
					Color:       color.RGBA{255, 255, 255, 255}, // 흰색으로 표시
					Trail:       make([]Vector3D, 0),
					MaxTrailLen: int(math.Max(float64(p1.MaxTrailLen), float64(p2.MaxTrailLen))),
				}

				// 기존 파티클들 제거하고 새 파티클 추가
				g.particles = append(g.particles[:i], g.particles[i+1:]...)
				if j > i {
					j--
				}
				g.particles = append(g.particles[:j], g.particles[j+1:]...)
				g.particles = append(g.particles, newParticle)

				// 선택된 인덱스 조정
				if g.selectedIndex == i || g.selectedIndex == j {
					g.selectedIndex = len(g.particles) - 1
				} else if g.selectedIndex > j {
					g.selectedIndex -= 2
				} else if g.selectedIndex > i {
					g.selectedIndex--
				}

				return // 한 번에 하나의 충돌만 처리
			}
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 20, 255}) // 어두운 파란색 배경

	// 시공간 그리드 그리기
	if g.showGrid {
		g.drawSpacetimeGrid(screen)
	}

	// 파티클 궤적 그리기
	g.drawTrails(screen)

	// 파티클 그리기
	g.drawParticles(screen)

	// UI 정보 표시
	g.drawUI(screen)
}

func (g *Game) drawSpacetimeGrid(screen *ebiten.Image) {
	for i := 0; i < g.grid.Height; i++ {
		for j := 0; j < g.grid.Width; j++ {
			curvature := g.grid.Curvature[i][j]

			if curvature > 0 {
				// 월드 좌표를 화면 좌표로 변환
				worldX := float64(j * gridSize)
				worldY := float64(i * gridSize)

				screenX := (worldX-g.camera.X-screenWidth/2)*g.zoom + screenWidth/2
				screenY := (worldY-g.camera.Y-screenHeight/2)*g.zoom + screenHeight/2

				// 곡률에 따른 색상 강도
				intensity := math.Min(curvature*1e12, 1.0)
				alpha := uint8(intensity * 100)

				if alpha > 5 {
					// 3D 효과를 위한 높이 계산
					height := curvature * 1e11

					// 그리드 포인트를 원으로 표시
					vector.DrawFilledCircle(screen, float32(screenX), float32(screenY-height), 2, color.RGBA{255, 0, 0, alpha}, false)

					// 연결선 그리기 (3D 효과)
					if j > 0 {
						prevWorldX := float64((j - 1) * gridSize)
						prevScreenX := (prevWorldX-g.camera.X-screenWidth/2)*g.zoom + screenWidth/2
						prevHeight := g.grid.Curvature[i][j-1] * 1e11

						vector.StrokeLine(screen, float32(prevScreenX), float32(screenY-prevHeight), float32(screenX), float32(screenY-height), 1, color.RGBA{100, 0, 0, alpha / 2}, false)
					}

					if i > 0 {
						prevWorldY := float64((i - 1) * gridSize)
						prevScreenY := (prevWorldY-g.camera.Y-screenHeight/2)*g.zoom + screenHeight/2
						prevHeight := g.grid.Curvature[i-1][j] * 1e11

						vector.StrokeLine(screen, float32(screenX), float32(prevScreenY-prevHeight), float32(screenX), float32(screenY-height), 1, color.RGBA{100, 0, 0, alpha / 2}, false)
					}
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

			// 월드 좌표를 화면 좌표로 변환
			screen1X := (pos1.X-g.camera.X-screenWidth/2)*g.zoom + screenWidth/2
			screen1Y := (pos1.Y-g.camera.Y-screenHeight/2)*g.zoom + screenHeight/2
			screen2X := (pos2.X-g.camera.X-screenWidth/2)*g.zoom + screenWidth/2
			screen2Y := (pos2.Y-g.camera.Y-screenHeight/2)*g.zoom + screenHeight/2

			// 궤적 색상 (점점 흐려짐)
			alpha := uint8(float64(i) / float64(len(p.Trail)) * 100)
			trailColor := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}

			vector.StrokeLine(screen, float32(screen1X), float32(screen1Y), float32(screen2X), float32(screen2Y), 1, trailColor, false)
		}
	}
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for i, p := range g.particles {
		// 월드 좌표를 화면 좌표로 변환
		screenX := (p.Position.X-g.camera.X-screenWidth/2)*g.zoom + screenWidth/2
		screenY := (p.Position.Y-g.camera.Y-screenHeight/2)*g.zoom + screenHeight/2

		// Z 좌표에 따른 3D 효과
		zOffset := p.Position.Z * 0.1
		screenY -= zOffset

		radius := float32(p.Radius * g.zoom)

		// 선택된 파티클 강조
		if i == g.selectedIndex {
			vector.StrokeCircle(screen, float32(screenX), float32(screenY), radius+3, 2, color.RGBA{255, 255, 255, 255}, false)
		}

		// 파티클 그리기
		vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), radius, p.Color, false)

		// 질량이 큰 파티클은 후광 효과
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
	// 상태 정보
	status := fmt.Sprintf("파티클: %d | 시간배율: %.1fx | 줌: %.1fx", len(g.particles), g.timeScale, g.zoom)
	if g.paused {
		status += " | 일시정지"
	}
	ebitenutil.DebugPrint(screen, status)

	// 조작법
	instructions := []string{
		"조작법:",
		"SPACE: 일시정지/재생",
		"G: 그리드 표시 토글",
		"+/-: 시간 배율 조정",
		"Z/X: 줌 인/아웃",
		"화살표: 카메라 이동",
		"좌클릭: 파티클 추가",
		"우클릭: 파티클 선택",
	}

	for i, instruction := range instructions {
		ebitenutil.DebugPrintAt(screen, instruction, 10, 30+i*15)
	}

	// 선택된 파티클 정보
	if g.selectedIndex >= 0 && g.selectedIndex < len(g.particles) {
		p := g.particles[g.selectedIndex]
		info := fmt.Sprintf("선택된 파티클:\n질량: %.2e\n속도: %.1f\n위치: (%.1f, %.1f, %.1f)",
			p.Mass, p.Velocity.Magnitude(), p.Position.X, p.Position.Y, p.Position.Z)
		ebitenutil.DebugPrintAt(screen, info, screenWidth-200, 30)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("일반상대성이론 시뮬레이션 - N-Body 중력 시스템")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
