# (General) Relativity Simulator
A simple simulator designed to visualize the principles of general relativity, written in Go.

This project aims to create a visual representation of how gravity works according to the theory of general relativity.

Currently, all physics calculations are performed on the CPU. Future plans include offloading these computations to the GPU using GLSL for a significant performance boost.

## Getting Started
There are two ways to run the simulator: downloading a pre-built release or building it from the source code.

### Option 1: Download a Release
Pre-compiled binaries for Windows (x64) are available on the [GitHub Releases page](https://github.com/deveworld/relativity_simul/releases). This is the quickest way to get started.

### Option 2: Build from Source
If you are using an operating system other than Windows, or if you simply prefer to compile it yourself, you can build the project from its source code.

First, clone the repository to your local machine:
```
git clone https://github.com/deveworld/relativity_simul
cd relativity_simul
```

Then, use the Go toolchain to build the executable:
```
go build .
```

## Controls
Once the simulation is running, you can navigate the camera using your keyboard.
- W, A, S, D: Move forward, left, backward, and right.
- Q, E: Move up and down.

Note: The ability to adjust simulation parameters in real-time is not yet implemented.
