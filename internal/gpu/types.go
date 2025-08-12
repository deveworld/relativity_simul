package gpu

// GPU holds the GPU context and state
type GPU struct {
	Initialized  bool
	Headless     bool
	NeedsCleanup bool                      // True if we need to clean up raylib context
	FftPlanCache map[string]*GPUFFTPlan    // Cache FFT plans by size/direction
	ShaderCache  map[string]*ComputeShader // Cache compiled shaders by source
}

// GPUMemoryBuffer represents a GPU memory buffer
type GPUMemoryBuffer struct {
	BufferID uint32
	Size     int
}

// ComputeShader represents a compute shader program
type ComputeShader struct {
	ProgramID uint32
}

// GPUFFTPlan holds FFT plan for GPU execution
type GPUFFTPlan struct {
	Gpu       *GPU
	Width     int
	Height    int
	IsForward bool
}

// ComplexGPUBuffer represents a GPU buffer for complex numbers
type ComplexGPUBuffer struct {
	BufferID uint32
	Size     int // Number of complex elements
}
