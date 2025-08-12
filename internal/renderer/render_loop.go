package renderer

import (
	"time"
)

// RenderLoop manages the main render loop
type RenderLoop struct {
	// Frame rate control
	targetFPS       int
	targetFrameTime float64
	lastFrameTime   float64
	actualFPS       int

	// State
	running      bool
	shouldClose  bool
	vsyncEnabled bool

	// Statistics
	frameCount      int
	frameTimes      []float64
	maxFrameSamples int

	// Callbacks
	updateCallback func(dt float64)
	renderCallback func(dt float64)
	beginCallback  func()
	endCallback    func()

	// Timing
	frameStartTime time.Time
}

// NewRenderLoop creates a new render loop
func NewRenderLoop() *RenderLoop {
	loop := &RenderLoop{
		targetFPS:       60, // Default from main.go
		maxFrameSamples: 60,
		frameTimes:      make([]float64, 0, 60),
	}
	loop.targetFrameTime = 1.0 / float64(loop.targetFPS)
	return loop
}

// GetTargetFPS returns the target FPS
func (r *RenderLoop) GetTargetFPS() int {
	return r.targetFPS
}

// SetTargetFPS sets the target FPS
func (r *RenderLoop) SetTargetFPS(fps int) {
	r.targetFPS = fps
	r.targetFrameTime = 1.0 / float64(fps)
}

// GetTargetFrameTime returns the target frame time in seconds
func (r *RenderLoop) GetTargetFrameTime() float64 {
	return r.targetFrameTime
}

// RecordFrameTime records a frame time for statistics
func (r *RenderLoop) RecordFrameTime(frameTime float64) {
	r.lastFrameTime = frameTime

	// Add to statistics
	r.frameTimes = append(r.frameTimes, frameTime)
	if len(r.frameTimes) > r.maxFrameSamples {
		r.frameTimes = r.frameTimes[1:]
	}

	// Calculate actual FPS
	if frameTime > 0 {
		r.actualFPS = int(1.0 / frameTime)
	}
}

// GetLastFrameTime returns the last recorded frame time
func (r *RenderLoop) GetLastFrameTime() float64 {
	return r.lastFrameTime
}

// GetActualFPS returns the actual FPS based on frame times
func (r *RenderLoop) GetActualFPS() int {
	return r.actualFPS
}

// IsRunning returns whether the render loop is running
func (r *RenderLoop) IsRunning() bool {
	return r.running
}

// Start starts the render loop
func (r *RenderLoop) Start() {
	r.running = true
	r.shouldClose = false
}

// Stop stops the render loop
func (r *RenderLoop) Stop() {
	r.running = false
}

// BeginFrame marks the beginning of a frame
func (r *RenderLoop) BeginFrame() {
	r.frameStartTime = time.Now()
	if r.beginCallback != nil {
		r.beginCallback()
	}
}

// EndFrame marks the end of a frame and enforces frame rate limit
func (r *RenderLoop) EndFrame() {
	if r.endCallback != nil {
		r.endCallback()
	}

	// Calculate elapsed time
	elapsed := time.Since(r.frameStartTime)

	// Wait to maintain target FPS
	targetDuration := time.Duration(r.targetFrameTime * float64(time.Second))
	if elapsed < targetDuration {
		time.Sleep(targetDuration - elapsed)
	}
}

// SetRenderCallback sets the render callback
func (r *RenderLoop) SetRenderCallback(callback func(dt float64)) {
	r.renderCallback = callback
}

// ExecuteFrame executes one frame with all callbacks
func (r *RenderLoop) ExecuteFrame() {
	// Use last frame time or default
	dt := r.lastFrameTime
	if dt == 0 {
		dt = r.targetFrameTime
	}

	// Begin phase
	if r.beginCallback != nil {
		r.beginCallback()
	}

	// Update phase
	if r.updateCallback != nil {
		r.updateCallback(dt)
	}

	// Render phase
	if r.renderCallback != nil {
		r.renderCallback(dt)
	}

	// End phase
	if r.endCallback != nil {
		r.endCallback()
	}

	r.frameCount++
}

// SetUpdateCallback sets the update callback
func (r *RenderLoop) SetUpdateCallback(callback func(dt float64)) {
	r.updateCallback = callback
}

// GetAverageFrameTime returns the average frame time
func (r *RenderLoop) GetAverageFrameTime() float64 {
	if len(r.frameTimes) == 0 {
		return r.targetFrameTime
	}

	sum := 0.0
	for _, ft := range r.frameTimes {
		sum += ft
	}
	return sum / float64(len(r.frameTimes))
}

// GetFrameCount returns the total frame count
func (r *RenderLoop) GetFrameCount() int {
	return len(r.frameTimes)
}

// IsVSyncEnabled returns whether vsync is enabled
func (r *RenderLoop) IsVSyncEnabled() bool {
	return r.vsyncEnabled
}

// EnableVSync enables or disables vsync
func (r *RenderLoop) EnableVSync(enable bool) {
	r.vsyncEnabled = enable
}

// SetBeginCallback sets the begin frame callback
func (r *RenderLoop) SetBeginCallback(callback func()) {
	r.beginCallback = callback
}

// SetEndCallback sets the end frame callback
func (r *RenderLoop) SetEndCallback(callback func()) {
	r.endCallback = callback
}

// ShouldClose returns whether the loop should close
func (r *RenderLoop) ShouldClose() bool {
	return r.shouldClose
}

// RequestClose requests the loop to close
func (r *RenderLoop) RequestClose() {
	r.shouldClose = true
}

// ResetCloseRequest resets the close request
func (r *RenderLoop) ResetCloseRequest() {
	r.shouldClose = false
}

// Run runs the main render loop
func (r *RenderLoop) Run() {
	r.Start()

	for r.running && !r.shouldClose {
		frameStart := time.Now()

		// Execute frame
		r.ExecuteFrame()

		// Record frame time
		frameTime := time.Since(frameStart).Seconds()
		r.RecordFrameTime(frameTime)

		// Frame rate limiting
		if !r.vsyncEnabled {
			elapsed := time.Since(frameStart)
			targetDuration := time.Duration(r.targetFrameTime * float64(time.Second))
			if elapsed < targetDuration {
				time.Sleep(targetDuration - elapsed)
			}
		}
	}

	r.Stop()
}

// GetStatistics returns render statistics
func (r *RenderLoop) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"targetFPS":        r.targetFPS,
		"actualFPS":        r.actualFPS,
		"frameCount":       r.frameCount,
		"averageFrameTime": r.GetAverageFrameTime(),
		"lastFrameTime":    r.lastFrameTime,
		"vsyncEnabled":     r.vsyncEnabled,
	}
}
