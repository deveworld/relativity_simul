# 2D Relativity Simulation

A high-performance N-body gravitational simulation in (2+1)D spacetime with GPU acceleration support, written in Go. This simulation implements weak-field general relativity approximations using the Particle-Mesh (PM) method with Fast Fourier Transform (FFT) acceleration.

## Features

- **Real-time N-body simulation** with gravitational interactions
- **GPU acceleration** using OpenGL 4.3+ compute shaders
- **Weak-field General Relativity** approximation in (2+1)D spacetime
- **Particle-Mesh (PM) method** for efficient force calculations
- **FFT-based Poisson solver** for gravitational potential
- **Interactive 3D visualization** with deformable spacetime grid
- **Dynamic camera controls** for exploration
- **Automatic CPU fallback** when GPU is unavailable

## Technology Stack

- **Language**: Go 1.24.3
- **Graphics**: [raylib-go](https://github.com/gen2brain/raylib-go) - 3D rendering and window management
- **GPU Computing**: [go-gl](https://github.com/go-gl/gl) - OpenGL 4.3+ compute shaders
- **FFT Processing**: [go-dsp](https://github.com/mjibson/go-dsp) - Digital signal processing
- **Testing**: [testify](https://github.com/stretchr/testify) - Testing toolkit

## Architecture

### Core Components

- **Physics Engine** (`internal/physics/`)
  - Particle dynamics with position and velocity
  - Force calculations using PM method
  - Time evolution with Kick-Drift-Kick integrator
  - Mass density grid deposition (Cloud-in-Cell)
  - Gradient computation for acceleration fields

- **GPU Acceleration** (`internal/gpu/`)
  - OpenGL compute shader management
  - FFT implementation (Cooley-Tukey for power-of-2, naive DFT fallback)
  - Buffer management for GPU memory
  - Automatic fallback to CPU on GPU errors

- **Rendering System** (`internal/renderer/`)
  - 3D particle visualization
  - Deformable spacetime grid representation
  - Camera controls and navigation
  - UI overlay with simulation stats

- **Input Handling** (`internal/input/`)
  - Mouse-based camera rotation
  - Keyboard controls for movement
  - Simulation control (pause, GPU toggle)

## Installation

### Prerequisites

- Go 1.24.3 or later
- OpenGL 4.3+ compatible GPU (optional, for GPU acceleration)
- C compiler (for CGO dependencies)
- Linux/Windows/macOS with OpenGL support

### Linux Dependencies

```bash
# Ubuntu/Debian
sudo apt-get install libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install mesa-libGL-devel libX11-devel libXrandr-devel libXinerama-devel libXcursor-devel libXi-devel libXxf86vm-devel
```

### macOS Dependencies

```bash
# Install Xcode Command Line Tools
xcode-select --install
```

### Windows Dependencies

- Visual Studio Build Tools or MinGW-w64
- No additional packages required

### Build and Run

```bash
# Clone the repository
git clone <repository-url>
cd relativity_simul_2d

# Download dependencies
go mod download

# Build the application
go build -o relativity_simulation

# Run the simulation
./relativity_simulation

# Or use the Makefile
make run
```

## Usage

### Controls

- **Camera Movement**
  - `Right Mouse Button + Move`: Look around
  - `W/S`: Move forward/backward
  - `A/D`: Move left/right
  - `Q/E`: Move up/down

- **Simulation Control**
  - `P`: Pause/unpause simulation
  - `G`: Toggle GPU/CPU mode
  - `ESC`: Exit application

### Configuration

The simulation parameters can be modified in `internal/config/config.go`:

```go
// Display settings
ScreenWidth:  1920,
ScreenHeight: 1080,

// Simulation dimensions
SimulationWidth: 256,  // Grid width
SimulationDepth: 256,  // Grid depth

// Physics parameters
NumParticles:          10,
GravitationalConstant: 1.0,

// Rendering parameters
GridVisScale:     10.0,
MoveSpeed:        5.0,
MouseSensitivity: 0.005,

// Runtime flags
StartPaused: false,
UseGPU:      true,
```

## Development

### Project Structure

```
relativity_simul_2d/
├── main.go                 # Application entry point and core simulation loop
├── Makefile               # Build commands
├── go.mod                 # Go module definition
├── internal/
│   ├── config/           # Configuration management
│   ├── gpu/              # GPU acceleration and compute shaders
│   ├── input/            # Input handling (keyboard, mouse)
│   ├── physics/          # Physics engine and calculations
│   ├── renderer/         # 3D rendering and visualization
│   └── simulation/       # Simulation state management
├── pkg/
│   └── fft/              # FFT implementations (CPU and GPU)
└── tests/
    └── integration/      # Integration and benchmark tests
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./tests/integration/

# Using Makefile
make test
```

### Code Quality

```bash
# Format code
make format

# Run linter (requires golangci-lint)
make check
```

## Physics Implementation

### Weak-Field General Relativity

The simulation implements a weak-field approximation of General Relativity in (2+1)D spacetime, where:

- The metric perturbation h₀₀ is related to the Newtonian potential Φ
- The field equation reduces to the Poisson equation: ∇²Φ = 4πGρ
- Particles follow geodesics approximated by Newtonian dynamics with relativistic corrections

### Particle-Mesh (PM) Method

1. **Mass Deposition**: Particles masses are deposited onto a regular grid using Cloud-in-Cell (CIC) interpolation
2. **Potential Calculation**: The Poisson equation is solved using FFT methods
3. **Force Calculation**: Forces are computed from the gradient of the potential
4. **Particle Update**: Particles are evolved using a Kick-Drift-Kick (KDK) integrator

### GPU Acceleration

The GPU implementation uses OpenGL compute shaders for:

- **FFT Operations**: Cooley-Tukey algorithm for power-of-2 sizes
- **Green's Function**: Applied in Fourier space for Poisson solving
- **Parallel Processing**: Efficient computation of grid operations

## Performance

### Benchmarks

The simulation achieves the following performance (example metrics):

- **CPU Mode**: ~60 FPS with 10 particles on 256x256 grid
- **GPU Mode**: ~60 FPS with 100+ particles on 256x256 grid
- **FFT Performance**: O(N log N) for power-of-2 sizes, O(N²) fallback for others

### Optimization Features

- Cached FFT plans for repeated transformations
- Shader compilation caching
- Efficient buffer management with ping-pong operations
- Automatic CPU fallback on GPU errors
- Frame-rate independent physics timestep

## Troubleshooting

### GPU Issues

- **"OpenGL context not available"**: Ensure your GPU supports OpenGL 4.3+
- **Automatic CPU fallback**: The simulation automatically falls back to CPU if GPU initialization fails
- **Performance degradation**: Check if GPU fallback is active (yellow indicator in UI)

### Build Issues

- **CGO errors**: Ensure C compiler is properly installed
- **OpenGL headers missing**: Install development packages for your OS
- **Module errors**: Run `go mod tidy` to resolve dependencies

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Follow TDD principles
2. Write tests for new features
3. Run `make format` and `make check` before submitting
4. Keep commits atomic and well-described
5. Update documentation as needed

## License

Apache License 2.0

## Acknowledgments

- Raylib for the excellent graphics library
- The Go community for the robust ecosystem
- OpenGL for GPU compute capabilities

## References

- [Particle-Mesh Methods](https://en.wikipedia.org/wiki/Particle_mesh)
- [Weak-field approximation in GR](https://en.wikipedia.org/wiki/Linearized_gravity)
- [Fast Fourier Transform](https://en.wikipedia.org/wiki/Fast_Fourier_transform)
- [OpenGL Compute Shaders](https://www.khronos.org/opengl/wiki/Compute_Shader)
