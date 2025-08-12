package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/go-gl/gl/v4.3-core/gl"
	"math"
	"relativity_simulation_2d/internal/config"
	"relativity_simulation_2d/internal/gpu"
	"relativity_simulation_2d/internal/input"
	"relativity_simulation_2d/internal/physics"
	"time"
)

var (
	cfg              *config.Config
	pause            bool
	useGPU           bool
	mouseSensitivity float32
	yaw              float32
	pitch            float32
)

// Simulation holds the entire state of the GR simulation
type Simulation struct {
	Particles       []*physics.Particle
	PotentialGrid   [][]float64 // Stores the scalar potential Φ (proportional to h_00)
	MassDensityGrid [][]float64 // Stores the mass density ρ
	AccelFieldX     [][]float64 // Stores the X component of the acceleration field
	AccelFieldZ     [][]float64 // Stores the Z component of the acceleration field
	gpu             *gpu.GPU    // Optional GPU context for acceleration (nil = CPU-only)

	// Error handling state for testing
	forceGPUInitFailure bool // For testing GPU initialization failures
	forceGPUCompFailure bool // For testing GPU computation failures
	gpuErrorOccurred    bool // Tracks if GPU error occurred
	fallbackToCPU       bool // Tracks if fallback to CPU was triggered
}

// NewSimulation creates and initializes a new simulation instance
func NewSimulation() *Simulation {
	sim := &Simulation{
		Particles:       make([]*physics.Particle, cfg.NumParticles),
		PotentialGrid:   make([][]float64, cfg.SimulationWidth),
		MassDensityGrid: make([][]float64, cfg.SimulationWidth),
		AccelFieldX:     make([][]float64, cfg.SimulationWidth),
		AccelFieldZ:     make([][]float64, cfg.SimulationWidth),
	}

	for i := range sim.PotentialGrid {
		sim.PotentialGrid[i] = make([]float64, cfg.SimulationDepth)
		sim.MassDensityGrid[i] = make([]float64, cfg.SimulationDepth)
		sim.AccelFieldX[i] = make([]float64, cfg.SimulationDepth)
		sim.AccelFieldZ[i] = make([]float64, cfg.SimulationDepth)
	}

	// Initialize particles using extracted function
	sim.Particles = physics.InitializeParticles(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth))

	// Optionally add a large central mass (uncomment to enable)
	// sim.Particles = physics.InitializeParticlesWithCentralMass(cfg.NumParticles, float64(cfg.SimulationWidth), float64(cfg.SimulationDepth), 1000)

	return sim
}

// CleanupGPU releases GPU resources if allocated
func (s *Simulation) CleanupGPU() {
	if s.gpu != nil {
		_ = CleanupGPU(s.gpu) // Ignore cleanup errors
		s.gpu = nil
	}
}

// HasGPUErrorOccurred returns true if a GPU error was encountered
func (s *Simulation) HasGPUErrorOccurred() bool {
	return s.gpuErrorOccurred
}

// Update runs one full step of the simulation with frame-rate independent timing
func (s *Simulation) Update(deltaTime float32) {
	// Use the extracted physics engine for time evolution
	forceField := physics.RunTimeEvolution(s.Particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)

	// Update our internal acceleration fields for visualization
	s.AccelFieldX = forceField.AccelFieldX
	s.AccelFieldZ = forceField.AccelFieldZ

	// Update mass density grid for visualization
	s.MassDensityGrid = physics.DepositMassToGrid(s.Particles, cfg.SimulationWidth, cfg.SimulationDepth)

	// Update potential grid for visualization
	s.PotentialGrid = physics.SolvePoissonFFT(s.MassDensityGrid, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
}

// solvePotential solves ∇²Φ = 4πGρ using FFT (kept for GPU fallback)
func (s *Simulation) solvePotential() {
	s.PotentialGrid = physics.SolvePoissonFFT(s.MassDensityGrid, cfg.SimulationWidth, cfg.SimulationDepth, cfg.GravitationalConstant)
}

// Real GPU Types and Functions for OpenGL 4.3+ Compute Shaders
// These replace fake CPU implementations with actual GPU acceleration
// Uses raylib's OpenGL context instead of separate GLFW window

// isHeadlessEnvironment detects if we're running in a headless environment (tests)
func isHeadlessEnvironment() bool {
	// Check if raylib window is initialized
	return !rl.IsWindowReady()
}

// InitializeGPU initializes GPU with proper context handling
func InitializeGPU() (*gpu.GPU, error) {
	return InitializeGPUWithMode(false)
}

// InitializeGPUWithMode initializes GPU with optional headless mode
func InitializeGPUWithMode(forceHeadless bool) (*gpu.GPU, error) {
	// Check if we need to create a headless context
	headless := forceHeadless || isHeadlessEnvironment()

	var needsCleanup bool
	if headless {
		// Initialize minimal raylib context for OpenGL
		rl.SetConfigFlags(rl.FlagWindowHidden)
		rl.InitWindow(1, 1, "GPU Test Context")
		needsCleanup = true
	}

	// Initialize OpenGL function pointers
	if err := gl.Init(); err != nil {
		if needsCleanup {
			rl.CloseWindow()
		}
		return nil, fmt.Errorf("failed to initialize OpenGL: %v", err)
	}

	// Test if OpenGL context is working
	var testBuffer uint32
	gl.GenBuffers(1, &testBuffer)
	if testBuffer == 0 {
		glError := gl.GetError()
		if needsCleanup {
			rl.CloseWindow()
		}
		return nil, fmt.Errorf("OpenGL context not available: GenBuffers failed (GL error: %d)", glError)
	}
	gl.DeleteBuffers(1, &testBuffer)

	return &gpu.GPU{
		Initialized:  true,
		Headless:     headless,
		NeedsCleanup: needsCleanup,
		FftPlanCache: make(map[string]*gpu.GPUFFTPlan),
		ShaderCache:  make(map[string]*gpu.ComputeShader),
	}, nil
}

func AllocateGPUMemory(g *gpu.GPU, sizeBytes int) (*gpu.GPUMemoryBuffer, error) {
	if !g.Initialized {
		return nil, fmt.Errorf("GPU context not initialized")
	}

	var bufferID uint32
	gl.GenBuffers(1, &bufferID)

	// Debug: Check if GenBuffers is working
	if bufferID == 0 {
		// Try to get more info about OpenGL state
		glError := gl.GetError()
		return nil, fmt.Errorf("gl.GenBuffers returned 0, GL error: %d", glError)
	}

	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, bufferID)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, sizeBytes, gl.Ptr(nil), gl.DYNAMIC_DRAW)

	if glError := gl.GetError(); glError != gl.NO_ERROR {
		return nil, fmt.Errorf("OpenGL error during buffer allocation: %d", glError)
	}

	return &gpu.GPUMemoryBuffer{BufferID: bufferID, Size: sizeBytes}, nil
}

func CompileComputeShader(g *gpu.GPU, source string) (*gpu.ComputeShader, error) {
	if !g.Initialized {
		return nil, fmt.Errorf("GPU context not initialized")
	}

	// Create and compile compute shader
	shaderID := gl.CreateShader(gl.COMPUTE_SHADER)
	cSources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shaderID, 1, cSources, nil)
	free()
	gl.CompileShader(shaderID)

	// Check compilation status
	var status int32
	gl.GetShaderiv(shaderID, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shaderID, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength+1)
		gl.GetShaderInfoLog(shaderID, logLength, nil, &log[0])

		gl.DeleteShader(shaderID)
		return nil, fmt.Errorf("compute shader compilation failed: %s", string(log))
	}

	// Create program and link
	programID := gl.CreateProgram()
	gl.AttachShader(programID, shaderID)
	gl.LinkProgram(programID)

	// Check linking status
	gl.GetProgramiv(programID, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(programID, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength+1)
		gl.GetProgramInfoLog(programID, logLength, nil, &log[0])

		gl.DeleteProgram(programID)
		gl.DeleteShader(shaderID)
		return nil, fmt.Errorf("compute shader linking failed: %s", string(log))
	}

	// Clean up shader (program retains copy)
	gl.DeleteShader(shaderID)

	return &gpu.ComputeShader{ProgramID: programID}, nil
}

func DeleteComputeShader(shader *gpu.ComputeShader) error {
	if shader.ProgramID != 0 {
		gl.DeleteProgram(shader.ProgramID)
		shader.ProgramID = 0
	}
	return nil
}

func CreateGPUFFTPlan2D(g *gpu.GPU, width, height int, isForward bool) (*gpu.GPUFFTPlan, error) {
	if !g.Initialized {
		return nil, fmt.Errorf("GPU context not initialized")
	}

	// For now, implement as a Cooley-Tukey FFT using compute shaders
	// This is a placeholder - real implementation would use optimized GPU FFT library
	plan := &gpu.GPUFFTPlan{
		Gpu:       g,
		Width:     width,
		Height:    height,
		IsForward: isForward,
	}

	// Create compute shaders for FFT operations (simplified for TDD)
	// Real implementation would have bit-reversal, butterfly operations, and transpose shaders
	return plan, nil
}

func CreateComplexGPUBuffer(g *gpu.GPU, elementCount int) (*gpu.ComplexGPUBuffer, error) {
	if !g.Initialized {
		return nil, fmt.Errorf("GPU context not initialized")
	}

	// Complex data requires 2x space (real + imaginary as float32 pairs)
	sizeBytes := elementCount * 8 // 2 * sizeof(float32)

	// Use the same method that works for regular buffers
	memBuffer, err := AllocateGPUMemory(g, sizeBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate GPU memory for complex buffer: %v", err)
	}

	// Wrap the GPUMemoryBuffer in a ComplexGPUBuffer
	return &gpu.ComplexGPUBuffer{BufferID: memBuffer.BufferID, Size: elementCount}, nil
}

func UploadComplexData(buffer *gpu.ComplexGPUBuffer, data []complex128) error {
	if buffer.BufferID == 0 {
		return fmt.Errorf("invalid complex GPU buffer")
	}

	if len(data) > buffer.Size {
		return fmt.Errorf("data too large for buffer: %d > %d", len(data), buffer.Size)
	}

	// Convert complex128 to interleaved float32 (real, imag, real, imag, ...)
	float32Data := make([]float32, len(data)*2)
	for i, c := range data {
		float32Data[i*2] = float32(real(c))
		float32Data[i*2+1] = float32(imag(c))
	}

	expectedSize := len(float32Data) * 4 // float32 = 4 bytes
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, buffer.BufferID)
	gl.BufferSubData(gl.SHADER_STORAGE_BUFFER, 0, expectedSize, gl.Ptr(float32Data))
	gl.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT)

	if glError := gl.GetError(); glError != gl.NO_ERROR {
		return fmt.Errorf("OpenGL error during complex data upload: %d", glError)
	}

	return nil
}

func DownloadComplexData(buffer *gpu.ComplexGPUBuffer, elementCount int) ([]complex128, error) {
	if buffer.BufferID == 0 {
		return nil, fmt.Errorf("invalid complex GPU buffer")
	}

	if elementCount > buffer.Size {
		return nil, fmt.Errorf("requested size too large: %d > %d", elementCount, buffer.Size)
	}

	// Download as interleaved float32, then convert to complex128
	float32Data := make([]float32, elementCount*2)
	expectedSize := len(float32Data) * 4 // float32 = 4 bytes

	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, buffer.BufferID)
	gl.GetBufferSubData(gl.SHADER_STORAGE_BUFFER, 0, expectedSize, gl.Ptr(float32Data))

	if glError := gl.GetError(); glError != gl.NO_ERROR {
		return nil, fmt.Errorf("OpenGL error during complex data download: %d", glError)
	}

	// Convert float32 pairs back to complex128
	data := make([]complex128, elementCount)
	for i := 0; i < elementCount; i++ {
		realData := float64(float32Data[i*2])
		imagData := float64(float32Data[i*2+1])
		data[i] = complex(realData, imagData)
	}

	return data, nil
}

func FreeComplexGPUBuffer(buffer *gpu.ComplexGPUBuffer) error {
	if buffer.BufferID != 0 {
		gl.DeleteBuffers(1, &buffer.BufferID)
		buffer.BufferID = 0
	}
	return nil
}

func ExecuteFFT(plan *gpu.GPUFFTPlan, inputBuffer, outputBuffer *gpu.ComplexGPUBuffer) error {
	if inputBuffer == nil || outputBuffer == nil {
		return fmt.Errorf("input and output buffers must not be nil")
	}

	totalSize := plan.Width * plan.Height

	// Create FFT shader for this execution
	fftShader, err := compileFFTComputeShader(plan.Width, plan.Height, plan.IsForward)
	if err != nil {
		return fmt.Errorf("failed to compile FFT shader: %v", err)
	}
	defer func() {
		_ = DeleteComputeShader(fftShader) // Ignore error during cleanup
	}()

	// Check if we're using Cooley-Tukey (power of 2) or fallback naive DFT
	if !isPowerOfTwo(plan.Width) || !isPowerOfTwo(plan.Height) {
		// Naive DFT - single pass
		return executeNaiveFFT(plan, fftShader, inputBuffer, outputBuffer, totalSize)
	}

	// Try Cooley-Tukey FFT - multi-stage execution
	err = executeCooleyTukeyFFT(plan, fftShader, inputBuffer, outputBuffer)
	if err != nil {
		// Fallback to naive DFT if Cooley-Tukey implementation is incomplete
		// This allows progressive implementation while maintaining functionality
		_ = DeleteComputeShader(fftShader) // Clean up the Cooley-Tukey shader

		// Create naive DFT shader for fallback
		naiveFftShader, naiveErr := compileNaiveDFTShader(plan.Width, plan.Height, plan.IsForward)
		if naiveErr != nil {
			return fmt.Errorf("Cooley-Tukey failed (%v) and naive fallback failed (%v)", err, naiveErr)
		}
		defer func() {
			_ = DeleteComputeShader(naiveFftShader)
		}()

		return executeNaiveFFT(plan, naiveFftShader, inputBuffer, outputBuffer, totalSize)
	}

	return nil
}

func executeNaiveFFT(plan *gpu.GPUFFTPlan, shader *gpu.ComputeShader, inputBuffer, outputBuffer *gpu.ComplexGPUBuffer, totalSize int) error {
	// Single-pass naive DFT
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, inputBuffer.BufferID)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 1, outputBuffer.BufferID)

	gl.UseProgram(shader.ProgramID)

	workGroupsX := uint32((totalSize + 63) / 64)
	gl.DispatchCompute(workGroupsX, 1, 1)
	gl.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT)

	if glError := gl.GetError(); glError != gl.NO_ERROR {
		return fmt.Errorf("OpenGL error during naive FFT execution: %d", glError)
	}

	return nil
}

func executeCooleyTukeyFFT(plan *gpu.GPUFFTPlan, shader *gpu.ComputeShader, inputBuffer, outputBuffer *gpu.ComplexGPUBuffer) error {
	gl.UseProgram(shader.ProgramID)

	// Get uniform locations
	stageLocation := gl.GetUniformLocation(shader.ProgramID, gl.Str("stage\x00"))
	directionLocation := gl.GetUniformLocation(shader.ProgramID, gl.Str("direction_flag\x00"))
	columnPassLocation := gl.GetUniformLocation(shader.ProgramID, gl.Str("is_column_pass\x00"))

	direction := int32(1)
	if !plan.IsForward {
		direction = -1
	}
	gl.Uniform1i(directionLocation, direction)

	// Create temporary buffer for ping-pong operations
	tempBuffer, err := createTempComplexBuffer(plan, plan.Width*plan.Height)
	if err != nil {
		return fmt.Errorf("failed to create temp buffer: %v", err)
	}
	defer func() {
		_ = FreeComplexGPUBuffer(tempBuffer)
	}()

	currentInput := inputBuffer
	currentOutput := tempBuffer

	// Phase 1: Row-wise FFT
	gl.Uniform1i(columnPassLocation, 0) // Row pass
	totalSize := uint32(plan.Width * plan.Height)
	workGroups := (totalSize + 31) / 32

	// Row bit-reversal pass
	gl.Uniform1i(stageLocation, -1) // Special stage for a bit of reversal
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, currentInput.BufferID)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 1, currentOutput.BufferID)
	gl.DispatchCompute(workGroups, 1, 1)
	gl.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT)

	// Swap buffers
	currentInput, currentOutput = currentOutput, currentInput

	// Row butterfly stages
	numStages := int(math.Log2(float64(plan.Width)))
	for stage := 0; stage < numStages; stage++ {
		gl.Uniform1i(stageLocation, int32(stage))
		gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, currentInput.BufferID)
		gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 1, currentOutput.BufferID)
		gl.DispatchCompute(workGroups, 1, 1)
		gl.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT)

		// Swap buffers
		currentInput, currentOutput = currentOutput, currentInput
	}

	// Phase 2: Column-wise FFT
	gl.Uniform1i(columnPassLocation, 1) // Column pass

	// Column bit-reversal pass
	gl.Uniform1i(stageLocation, -1)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, currentInput.BufferID)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 1, currentOutput.BufferID)
	gl.DispatchCompute(workGroups, 1, 1)
	gl.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT)

	// Swap buffers
	currentInput, currentOutput = currentOutput, currentInput

	// Column butterfly stages
	numStages = int(math.Log2(float64(plan.Height)))
	for stage := 0; stage < numStages; stage++ {
		gl.Uniform1i(stageLocation, int32(stage))
		gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, currentInput.BufferID)
		gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 1, currentOutput.BufferID)
		gl.DispatchCompute(workGroups, 1, 1)
		gl.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT)

		// Swap buffers
		currentInput, currentOutput = currentOutput, currentInput
	}

	// Copy final result to output buffer (if needed)
	if currentInput != outputBuffer {
		err = copyComplexBuffer(plan, currentInput, outputBuffer)
		if err != nil {
			return fmt.Errorf("failed to copy final result: %v", err)
		}
	}

	if glError := gl.GetError(); glError != gl.NO_ERROR {
		return fmt.Errorf("OpenGL error during Cooley-Tukey FFT: %d", glError)
	}

	return nil
}

func createTempComplexBuffer(plan *gpu.GPUFFTPlan, elementCount int) (*gpu.ComplexGPUBuffer, error) {
	// Create a temporary complex buffer for intermediate FFT operations
	tempBuffer, err := CreateComplexGPUBuffer(plan.Gpu, elementCount)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary complex buffer: %v", err)
	}
	return tempBuffer, nil
}

func copyComplexBuffer(plan *gpu.GPUFFTPlan, src, dst *gpu.ComplexGPUBuffer) error {
	// Copy data between complex GPU buffers using OpenGL buffer copy
	if src.BufferID == 0 || dst.BufferID == 0 {
		return fmt.Errorf("invalid buffer IDs for copy operation")
	}

	// Calculate byte size (complex elements * 8 bytes per element)
	byteSize := src.Size * 8
	if dst.Size < src.Size {
		return fmt.Errorf("destination buffer too small: %d < %d", dst.Size, src.Size)
	}

	// Use OpenGL's efficient buffer-to-buffer copy
	gl.BindBuffer(gl.COPY_READ_BUFFER, src.BufferID)
	gl.BindBuffer(gl.COPY_WRITE_BUFFER, dst.BufferID)
	gl.CopyBufferSubData(gl.COPY_READ_BUFFER, gl.COPY_WRITE_BUFFER, 0, 0, byteSize)

	// Unbind buffers
	gl.BindBuffer(gl.COPY_READ_BUFFER, 0)
	gl.BindBuffer(gl.COPY_WRITE_BUFFER, 0)

	if glError := gl.GetError(); glError != gl.NO_ERROR {
		return fmt.Errorf("OpenGL error during buffer copy: %d", glError)
	}

	return nil
}

func DestroyFFTPlan(plan *gpu.GPUFFTPlan) error {
	// Clean up any allocated resources
	return nil
}

func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}

func compileNaiveDFTShader(width, height int, isForward bool) (*gpu.ComputeShader, error) {
	// Fallback to O(N²) DFT implementation for non-power-of-2 sizes
	direction := "1.0"
	if !isForward {
		direction = "-1.0"
	}

	shaderSource := fmt.Sprintf(`
		#version 430
		layout(local_size_x = 64) in;

		layout(std430, binding = 0) buffer InputBuffer {
			vec2 inputData[];
		};
		layout(std430, binding = 1) buffer OutputBuffer {
			vec2 outputData[];
		};

		const float PI = 3.14159265359;
		const float direction = %s;
		const int WIDTH = %d;
		const int HEIGHT = %d;
		const int TOTAL_SIZE = WIDTH * HEIGHT;

		vec2 complexMul(vec2 a, vec2 b) {
			return vec2(a.x * b.x - a.y * b.y, a.x * b.y + a.y * b.x);
		}

		void main() {
			uint index = gl_GlobalInvocationID.x;
			if (index >= TOTAL_SIZE) return;

			uint outputX = index %% WIDTH;
			uint outputY = index / WIDTH;

			vec2 sum = vec2(0.0, 0.0);

			for (uint inputY = 0; inputY < HEIGHT; inputY++) {
				for (uint inputX = 0; inputX < WIDTH; inputX++) {
					float angle = direction * 2.0 * PI * (
						float(outputX * inputX) / float(WIDTH) +
						float(outputY * inputY) / float(HEIGHT)
					);
					vec2 twiddle = vec2(cos(angle), sin(angle));
					vec2 inputSample = inputData[inputY * WIDTH + inputX];
					sum += complexMul(inputSample, twiddle);
				}
			}

			if (direction < 0.0) {
				float normFactor = 1.0 / float(TOTAL_SIZE);
				sum *= normFactor;
			}

			outputData[index] = sum;
		}
	`, direction, width, height)

	// Compile the compute shader
	shaderID := gl.CreateShader(gl.COMPUTE_SHADER)
	csources, free := gl.Strs(shaderSource + "\x00")
	gl.ShaderSource(shaderID, 1, csources, nil)
	free()
	gl.CompileShader(shaderID)

	// Check compilation status
	var status int32
	gl.GetShaderiv(shaderID, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shaderID, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength+1)
		gl.GetShaderInfoLog(shaderID, logLength, nil, &log[0])

		gl.DeleteShader(shaderID)
		return nil, fmt.Errorf("naive DFT compute shader compilation failed: %s", string(log))
	}

	// Create program and link
	programID := gl.CreateProgram()
	gl.AttachShader(programID, shaderID)
	gl.LinkProgram(programID)

	// Check linking status
	gl.GetProgramiv(programID, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(programID, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength+1)
		gl.GetProgramInfoLog(programID, logLength, nil, &log[0])

		gl.DeleteProgram(programID)
		gl.DeleteShader(shaderID)
		return nil, fmt.Errorf("naive DFT compute shader linking failed: %s", string(log))
	}

	// Clean up shader object (no longer needed after linking)
	gl.DeleteShader(shaderID)

	return &gpu.ComputeShader{ProgramID: programID}, nil
}

func compileFFTComputeShader(width, height int, isForward bool) (*gpu.ComputeShader, error) {
	// O(N log N) Cooley-Tukey FFT implementation for GPU
	// Uses separable 2D FFT: row FFTs then column FFTs

	// Check if dimensions are power of 2 (required for Cooley-Tukey)
	if !isPowerOfTwo(width) || !isPowerOfTwo(height) {
		return compileNaiveDFTShader(width, height, isForward)
	}

	shaderSource := fmt.Sprintf(`
		#version 430
		layout(local_size_x = 32, local_size_y = 1) in;

		layout(std430, binding = 0) buffer InputBuffer {
			vec2 inputData[];
		};
		layout(std430, binding = 1) buffer OutputBuffer {
			vec2 outputData[];
		};

		uniform int stage;      // Current FFT stage (0 to log2(size)-1)
		uniform int direction_flag;  // 1 for forward, -1 for inverse
		uniform int is_column_pass;  // 0 for row pass, 1 for column pass

		const float PI = 3.14159265359;
		const int WIDTH = %d;
		const int HEIGHT = %d;
		const int TOTAL_SIZE = WIDTH * HEIGHT;

		vec2 complexMul(vec2 a, vec2 b) {
			return vec2(a.x * b.x - a.y * b.y, a.x * b.y + a.y * b.x);
		}

		// Bit reversal for FFT
		uint bitReverse(uint x, uint bits) {
			uint result = 0;
			for (uint i = 0; i < bits; i++) {
				if ((x & (1u << i)) != 0) {
					result |= 1u << (bits - 1 - i);
				}
			}
			return result;
		}

		void main() {
			uint index = gl_GlobalInvocationID.x;

			if (is_column_pass == 0) {
				// Row pass: process each row independently
				uint row = index / WIDTH;
				uint col = index %% WIDTH;

				if (row >= HEIGHT || col >= WIDTH) return;

				if (stage == -1) {
					// Bit reversal stage for rows
					uint bits = uint(log2(float(WIDTH)));
					uint reversedCol = bitReverse(col, bits);
					uint srcIndex = row * WIDTH + col;
					uint dstIndex = row * WIDTH + reversedCol;
					outputData[dstIndex] = inputData[srcIndex];
				} else {
					// Butterfly operations for current stage
					uint stepSize = 1u << (stage + 1);
					uint halfStep = stepSize >> 1;
					uint group = col / stepSize;
					uint pos = col %% stepSize;

					if (pos < halfStep) {
						uint partner = index + halfStep;
						if (partner < TOTAL_SIZE) {
							float angle = float(direction_flag) * (-2.0 * PI * float(pos)) / float(stepSize);
							vec2 twiddle = vec2(cos(angle), sin(angle));

							vec2 a = inputData[index];
							vec2 b = complexMul(inputData[partner], twiddle);

							outputData[index] = a + b;
							outputData[partner] = a - b;
						}
					}
				}
			} else {
				// Column pass: process each column independently
				uint col = index / HEIGHT;
				uint row = index %% HEIGHT;

				if (col >= WIDTH || row >= HEIGHT) return;

				if (stage == -1) {
					// Bit reversal stage for columns
					uint bits = uint(log2(float(HEIGHT)));
					uint reversedRow = bitReverse(row, bits);
					uint srcIndex = row * WIDTH + col;
					uint dstIndex = reversedRow * WIDTH + col;
					outputData[dstIndex] = inputData[srcIndex];
				} else {
					// Butterfly operations for current stage
					uint stepSize = 1u << (stage + 1);
					uint halfStep = stepSize >> 1;
					uint group = row / stepSize;
					uint pos = row %% stepSize;

					if (pos < halfStep) {
						uint partnerRow = row + halfStep;
						if (partnerRow < HEIGHT) {
							uint currentIndex = row * WIDTH + col;
							uint partnerIndex = partnerRow * WIDTH + col;

							float angle = float(direction_flag) * (-2.0 * PI * float(pos)) / float(stepSize);
							vec2 twiddle = vec2(cos(angle), sin(angle));

							vec2 a = inputData[currentIndex];
							vec2 b = complexMul(inputData[partnerIndex], twiddle);

							outputData[currentIndex] = a + b;
							outputData[partnerIndex] = a - b;
						}
					}
				}
			}
		}
	`, width, height)

	// Compile the compute shader
	shaderID := gl.CreateShader(gl.COMPUTE_SHADER)
	csources, free := gl.Strs(shaderSource + "\x00")
	gl.ShaderSource(shaderID, 1, csources, nil)
	free()
	gl.CompileShader(shaderID)

	// Check compilation status
	var status int32
	gl.GetShaderiv(shaderID, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shaderID, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength+1)
		gl.GetShaderInfoLog(shaderID, logLength, nil, &log[0])

		gl.DeleteShader(shaderID)
		return nil, fmt.Errorf("FFT compute shader compilation failed: %s", string(log))
	}

	// Create program and link
	programID := gl.CreateProgram()
	gl.AttachShader(programID, shaderID)
	gl.LinkProgram(programID)

	// Check linking status
	gl.GetProgramiv(programID, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(programID, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength+1)
		gl.GetProgramInfoLog(programID, logLength, nil, &log[0])

		gl.DeleteProgram(programID)
		gl.DeleteShader(shaderID)
		return nil, fmt.Errorf("FFT compute shader linking failed: %s", string(log))
	}

	// Clean up shader (program retains copy)
	gl.DeleteShader(shaderID)

	return &gpu.ComputeShader{ProgramID: programID}, nil
}

func SolvePoissonGPU(g *gpu.GPU, densityGrid [][]float64, gravitationalConstant float64) ([][]float64, error) {
	if !g.Initialized {
		return nil, fmt.Errorf("GPU context not initialized")
	}

	// Match CPU coordinate system: densityGrid[i][j] where i=cfg.SimulationWidth, j=cfg.SimulationDepth
	width := len(densityGrid)     // cfg.SimulationWidth (first dimension)
	height := len(densityGrid[0]) // cfg.SimulationDepth (second dimension)
	totalSize := width * height

	// Step 1: Upload density grid to GPU as complex data (real part = density, imag = 0)
	inputBuffer, err := CreateComplexGPUBuffer(g, totalSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create input buffer: %v", err)
	}

	// Convert density to complex128 and upload (match CPU coordinate system)
	complexData := make([]complex128, totalSize)
	idx := 0
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			complexData[idx] = complex(densityGrid[i][j], 0) // Use i,j to match CPU reference
			idx++
		}
	}

	err = UploadComplexData(inputBuffer, complexData)
	if err != nil {
		return nil, fmt.Errorf("failed to upload density data: %v", err)
	}

	// Step 2: Forward FFT
	fftOutputBuffer, err := CreateComplexGPUBuffer(g, totalSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create FFT output buffer: %v", err)
	}

	// Step 2: Forward FFT (use cached plan if available)
	fftKey := fmt.Sprintf("%dx%d_fwd", width, height)
	fftPlan, exists := g.FftPlanCache[fftKey]
	if !exists {
		var err error
		fftPlan, err = CreateGPUFFTPlan2D(g, width, height, true) // forward FFT
		if err != nil {
			return nil, fmt.Errorf("failed to create FFT plan: %v", err)
		}
		g.FftPlanCache[fftKey] = fftPlan
	}

	err = ExecuteFFT(fftPlan, inputBuffer, fftOutputBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to execute forward FFT: %v", err)
	}

	// Step 3: Apply Green's function in Fourier space
	err = applyGreensFunction(g, fftOutputBuffer, width, height, gravitationalConstant)
	if err != nil {
		return nil, fmt.Errorf("failed to apply Green's function: %v", err)
	}

	// Step 4: Inverse FFT (use cached plan if available)
	ifftKey := fmt.Sprintf("%dx%d_inv", width, height)
	ifftPlan, exists := g.FftPlanCache[ifftKey]
	if !exists {
		var err error
		ifftPlan, err = CreateGPUFFTPlan2D(g, width, height, false) // inverse FFT
		if err != nil {
			return nil, fmt.Errorf("failed to create IFFT plan: %v", err)
		}
		g.FftPlanCache[ifftKey] = ifftPlan
	}

	finalBuffer, err := CreateComplexGPUBuffer(g, totalSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create final buffer: %v", err)
	}

	err = ExecuteFFT(ifftPlan, fftOutputBuffer, finalBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to execute inverse FFT: %v", err)
	}

	// Step 5: Download result and extract real part
	resultData, err := DownloadComplexData(finalBuffer, totalSize)
	if err != nil {
		return nil, fmt.Errorf("failed to download result: %v", err)
	}

	// Apply inverse FFT normalization to match CPU go-dsp library behavior
	// The CPU library auto-normalizes, but GPU IFFT may not
	normalizationFactor := 1.0 / float64(totalSize)
	for i := range resultData {
		resultData[i] *= complex(normalizationFactor, 0)
	}

	// Convert back to 2D real grid (match CPU coordinate system)
	potentialGrid := make([][]float64, width) // width rows (cfg.SimulationWidth)
	for i := 0; i < width; i++ {
		potentialGrid[i] = make([]float64, height) // height columns (cfg.SimulationDepth)
	}

	idx = 0
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			potentialGrid[i][j] = real(resultData[idx]) // Use i,j to match CPU coordinate system
			idx++
		}
	}

	return potentialGrid, nil
}

// applyGreensFunction applies Green's function kernel in Fourier space
func applyGreensFunction(g *gpu.GPU, buffer *gpu.ComplexGPUBuffer, width, height int, gravitationalConstant float64) error {
	// Create compute shader for Green's function
	shaderSource := fmt.Sprintf(`
		#version 430
		layout(local_size_x = 64) in;

		layout(std430, binding = 0) buffer FourierBuffer {
			vec2 fourierData[];
		};

		uniform int uWidth;
		uniform int uHeight;
		uniform float uGConstant;
		uniform float uKxFactor;
		uniform float uKzFactor;

		void main() {
			uint index = gl_GlobalInvocationID.x;
			uint totalSize = uint(uWidth * uHeight);

			if (index >= totalSize) return;

			// Convert 1D index to 2D coordinates
			// Data is uploaded as densityGrid[i][j] with j inner loop
			// So u (width/i) changes slower, v (height/j) changes faster
			uint u = index / uint(uHeight);
			uint v = index %% uint(uHeight);

			// Calculate wave vector k
			float kx = float(u);
			if (u > uint(uWidth)/2u) {
				kx = float(int(u) - uWidth);
			}

			float kz = float(v);
			if (v > uint(uHeight)/2u) {
				kz = float(int(v) - uHeight);
			}

			float kSquared = (kx * uKxFactor) * (kx * uKxFactor) +
							 (kz * uKzFactor) * (kz * uKzFactor);

			if (kSquared == 0.0) {
				// Ignore DC component
				fourierData[index] = vec2(0.0, 0.0);
			} else {
				// Apply Green's function: G(k) = -4πG / |k|²
				float scalingFactor = -4.0 * 3.14159265359 * uGConstant / kSquared;
				fourierData[index] *= scalingFactor;
			}
		}
	`)

	// Use cached shader if available
	shaderKey := "greens_function_shader"
	shader, exists := g.ShaderCache[shaderKey]
	if !exists {
		var err error
		shader, err = CompileComputeShader(g, shaderSource)
		if err != nil {
			return fmt.Errorf("failed to compile Green's function shader: %v", err)
		}
		g.ShaderCache[shaderKey] = shader
	}

	// Bind buffer and set uniforms
	gl.UseProgram(shader.ProgramID)

	// Bind buffer to shader storage buffer object
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, buffer.BufferID)

	// Set uniforms
	kxFactor := 2.0 * math.Pi / float64(width)
	kzFactor := 2.0 * math.Pi / float64(height)

	widthLoc := gl.GetUniformLocation(shader.ProgramID, gl.Str("uWidth\x00"))
	heightLoc := gl.GetUniformLocation(shader.ProgramID, gl.Str("uHeight\x00"))
	gravitationalConstantLoc := gl.GetUniformLocation(shader.ProgramID, gl.Str("uGConstant\x00"))
	kxFactorLoc := gl.GetUniformLocation(shader.ProgramID, gl.Str("uKxFactor\x00"))
	kzFactorLoc := gl.GetUniformLocation(shader.ProgramID, gl.Str("uKzFactor\x00"))

	gl.Uniform1i(widthLoc, int32(width))
	gl.Uniform1i(heightLoc, int32(height))
	gl.Uniform1f(gravitationalConstantLoc, float32(gravitationalConstant))
	gl.Uniform1f(kxFactorLoc, float32(kxFactor))
	gl.Uniform1f(kzFactorLoc, float32(kzFactor))

	// Dispatch compute shader
	totalSize := width * height
	workGroups := (totalSize + 63) / 64 // Round up to handle all elements
	gl.DispatchCompute(uint32(workGroups), 1, 1)

	// Memory barrier to ensure shader writes are visible
	gl.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT)

	// Check for OpenGL errors
	if glError := gl.GetError(); glError != gl.NO_ERROR {
		return fmt.Errorf("OpenGL error in Green's function shader: %d", glError)
	}

	return nil
}

func CleanupGPU(g *gpu.GPU) error {
	if g.Initialized {
		// Clean up cached FFT plans
		for _, plan := range g.FftPlanCache {
			_ = DestroyFFTPlan(plan)
		}
		g.FftPlanCache = nil

		// Clean up cached shaders
		for _, shader := range g.ShaderCache {
			_ = DeleteComputeShader(shader)
		}
		g.ShaderCache = nil

		// Clean up headless raylib context if we created one
		if g.NeedsCleanup {
			rl.CloseWindow()
		}

		g.Initialized = false
	}
	return nil
}

// UpdateGPU performs a simulation timestep using GPU acceleration for Poisson solver
func (s *Simulation) UpdateGPU(deltaTime float32) {
	// Use a hybrid approach: physics engine for particle updates, GPU for Poisson solver

	// 1. Kick (half step velocity update)
	forceField := &physics.ForceField{
		AccelFieldX: s.AccelFieldX,
		AccelFieldZ: s.AccelFieldZ,
		Width:       cfg.SimulationWidth,
		Height:      cfg.SimulationDepth,
	}
	forceCorrectionFactor := float32(0.5)
	physics.UpdateVelocities(s.Particles, forceField, deltaTime*0.5, forceCorrectionFactor)

	// 2. Drift (full step position update)
	physics.UpdatePositions(s.Particles, deltaTime, cfg.SimulationWidth, cfg.SimulationDepth)

	// 3. Calculate new accelerations using GPU-accelerated Poisson solver
	s.calculateAccelerationFieldGPU()

	// 4. Kick (half step velocity update)
	forceField.AccelFieldX = s.AccelFieldX
	forceField.AccelFieldZ = s.AccelFieldZ
	physics.UpdateVelocities(s.Particles, forceField, deltaTime*0.5, forceCorrectionFactor)
}

// calculateAccelerationFieldGPU performs PM method steps with GPU-accelerated potential calculation
func (s *Simulation) calculateAccelerationFieldGPU() {
	// Step 1: Deposit mass onto the grid (Cloud-in-Cell) - same as CPU
	s.MassDensityGrid = physics.DepositMassToGrid(s.Particles, cfg.SimulationWidth, cfg.SimulationDepth)

	// Step 2: Solve for potential Φ using GPU
	s.solvePotentialGPU()

	// Step 3: Calculate acceleration (a = -∇Φ) from the potential field
	forceField := physics.CalculateGradient(s.PotentialGrid, cfg.SimulationWidth, cfg.SimulationDepth)
	s.AccelFieldX = forceField.AccelFieldX
	s.AccelFieldZ = forceField.AccelFieldZ
}

// solvePotentialGPU solves ∇²Φ = 4πGρ using GPU-accelerated FFT
func (s *Simulation) solvePotentialGPU() {
	// Check for forced initialization failure (testing)
	if s.forceGPUInitFailure {
		s.gpuErrorOccurred = true
		s.fallbackToCPU = true
		s.solvePotential()
		return
	}

	// Initialize GPU context if needed
	if s.gpu == nil {
		GPU, err := InitializeGPU()
		if err != nil {
			// Fallback to CPU if GPU unavailable
			s.gpuErrorOccurred = true
			s.fallbackToCPU = true
			s.solvePotential()
			return
		}
		s.gpu = GPU
	}

	// Check for forced computation failure (testing)
	if s.forceGPUCompFailure {
		s.gpuErrorOccurred = true
		s.fallbackToCPU = true
		s.solvePotential()
		return
	}

	// Use GPU Poisson solver
	result, err := SolvePoissonGPU(s.gpu, s.MassDensityGrid, cfg.GravitationalConstant)
	if err != nil {
		// Fallback to CPU if GPU computation fails
		s.gpuErrorOccurred = true
		s.fallbackToCPU = true
		s.solvePotential()
		return
	}

	// Copy result back to PotentialGrid
	for i := range s.PotentialGrid {
		for j := range s.PotentialGrid[i] {
			s.PotentialGrid[i][j] = result[i][j]
		}
	}
}

func processInput(camera *rl.Camera3D) {
	// Process all input through the controller
	input.ProcessAllInput(camera, &pause, &useGPU, &yaw, &pitch, cfg.MoveSpeed, mouseSensitivity, int(cfg.ScreenWidth), int(cfg.ScreenHeight))
}

func main() {
	// Initialize configuration
	cfg = config.DefaultConfig()
	pause = cfg.StartPaused
	useGPU = cfg.UseGPU
	mouseSensitivity = cfg.MouseSensitivity
	yaw = cfg.InitialYaw
	pitch = cfg.InitialPitch

	// Initialize window
	rl.InitWindow(int32(cfg.ScreenWidth), int32(cfg.ScreenHeight), "Golang GR Simulation - (2+1)D Spacetime")
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
	defer simulation.CleanupGPU() // Clean up GPU resources on exit

	rl.HideCursor()
	rl.SetClipPlanes(0.1, 10000.0)
	rl.SetTargetFPS(60)
	// Main game loop
	for !rl.WindowShouldClose() {
		// Handle input
		processInput(&camera)

		// Update simulation state if not paused
		if !pause {
			// Use actual frame time for frame-rate independent simulation
			deltaTime := rl.GetFrameTime()
			// Cap delta time to prevent simulation instability during lag spikes
			if deltaTime > 0.05 {
				deltaTime = 0.05 // Max 20 FPS equivalent
			}

			start := time.Now()
			if useGPU {
				simulation.UpdateGPU(deltaTime) // Use GPU acceleration
			} else {
				simulation.Update(deltaTime)
			}
			_ = time.Since(start) // Measure simulation time (for future performance monitoring)
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
		rl.DrawSphere(p.Position.ToRaylib(), p.Radius, rl.Gold)
	}

	// Draw coordinate axes
	rl.DrawLine3D(rl.NewVector3(0, 0, 0), rl.NewVector3(5, 0, 0), rl.Red)   // X axis
	rl.DrawLine3D(rl.NewVector3(0, 0, 0), rl.NewVector3(0, 5, 0), rl.Green) // Y axis
	rl.DrawLine3D(rl.NewVector3(0, 0, 0), rl.NewVector3(0, 0, 5), rl.Blue)  // Z axis

	rl.EndMode3D()

	// Draw UI
	rl.DrawText("GR (Weak-Field) N-Body Simulation", 10, 10, 20, rl.Lime)
	rl.DrawText(fmt.Sprintf("Particles: %d", cfg.NumParticles), 10, 40, 20, rl.White)

	// GPU/CPU status indicator with GPU error status
	if useGPU {
		if sim.HasGPUErrorOccurred() {
			rl.DrawText("Mode: GPU (Fallback to CPU)", 10, 70, 20, rl.Yellow)
		} else {
			rl.DrawText("Mode: GPU Accelerated", 10, 70, 20, rl.Green)
		}
	} else {
		rl.DrawText("Mode: CPU Only", 10, 70, 20, rl.Orange)
	}

	rl.DrawText("Right-click + Mouse to look", 10, 130, 20, rl.White)
	rl.DrawText("W,A,S,D,Q,E to move", 10, 160, 20, rl.White)
	rl.DrawText("P to pause, G to toggle GPU", 10, 190, 20, rl.White)

	// Display both target and actual FPS
	targetFPS := 60
	actualFPS := rl.GetFPS()
	frameTime := rl.GetFrameTime()
	rl.DrawText(fmt.Sprintf("Target FPS: %d", targetFPS), int32(cfg.ScreenWidth)-200, 10, 20, rl.White)
	rl.DrawText(fmt.Sprintf("Actual FPS: %d", actualFPS), int32(cfg.ScreenWidth)-200, 35, 20, rl.White)
	rl.DrawText(fmt.Sprintf("Frame Time: %.3fs", frameTime), int32(cfg.ScreenWidth)-200, 60, 20, rl.White)

	if pause {
		rl.DrawText("PAUSED (Press P to unpause)", int32(cfg.ScreenWidth)/2-150, int32(cfg.ScreenHeight)/2-10, 20, rl.Yellow)
	}

	rl.EndDrawing()
}

func drawDeformedGrid(sim *Simulation) {
	gridColor := rl.NewColor(50, 50, 100, 255)

	// Draw lines parallel to Z axis
	for i := 0; i < cfg.SimulationWidth; i++ {
		for j := 0; j < cfg.SimulationDepth-1; j++ {
			p1X := float32(i) - float32(cfg.SimulationWidth)/2.0
			p1Z := float32(j) - float32(cfg.SimulationDepth)/2.0
			p1Y := float32(sim.PotentialGrid[i][j] * cfg.GridVisScale)

			p2X := float32(i) - float32(cfg.SimulationWidth)/2.0
			p2Z := float32(j+1) - float32(cfg.SimulationDepth)/2.0
			p2Y := float32(sim.PotentialGrid[i][j+1] * cfg.GridVisScale)

			rl.DrawLine3D(rl.NewVector3(p1X, p1Y, p1Z), rl.NewVector3(p2X, p2Y, p2Z), gridColor)
		}
	}

	// Draw lines parallel to X axis
	for j := 0; j < cfg.SimulationDepth; j++ {
		for i := 0; i < cfg.SimulationWidth-1; i++ {
			p1X := float32(i) - float32(cfg.SimulationWidth)/2.0
			p1Z := float32(j) - float32(cfg.SimulationDepth)/2.0
			p1Y := float32(sim.PotentialGrid[i][j] * cfg.GridVisScale)

			p2X := float32(i+1) - float32(cfg.SimulationWidth)/2.0
			p2Z := float32(j) - float32(cfg.SimulationDepth)/2.0
			p2Y := float32(sim.PotentialGrid[i+1][j] * cfg.GridVisScale)

			rl.DrawLine3D(rl.NewVector3(p1X, p1Y, p1Z), rl.NewVector3(p2X, p2Y, p2Z), gridColor)
		}
	}
}
