package physics

import (
	"math"
	"math/rand"
)

// InitializeParticles creates particles with random positions and masses
func InitializeParticles(numParticles int, simulationWidth, simulationDepth float64) []*Particle {
	particles := make([]*Particle, numParticles)

	for i := 0; i < numParticles; i++ {
		mass := 20.0 + rand.Float32()*30.0
		particles[i] = &Particle{
			Position: NewVec3(
				float64((rand.Float32()-0.5)*float32(simulationWidth)*0.8),
				0,
				float64((rand.Float32()-0.5)*float32(simulationDepth)*0.8),
			),
			Velocity: NewVec3(0, 0, 0),
			Mass:     mass,
			Radius:   float32(math.Pow(float64(mass/20.0), 1.0/3.0)) * 0.5,
		}
	}

	return particles
}

// InitializeParticlesWithCentralMass creates particles with a large central mass
func InitializeParticlesWithCentralMass(numParticles int, simulationWidth, simulationHeight float64, centralMass float64) []*Particle {
	particles := InitializeParticles(numParticles, simulationWidth, simulationHeight)

	// Replace first particle with central mass
	particles[0] = &Particle{
		Position: NewVec3(0, 0, 0),
		Velocity: NewVec3(0, 0, 0),
		Mass:     float32(centralMass),
		Radius:   float32(math.Pow(centralMass/20.0, 1.0/3.0)) * 0.5,
	}

	return particles
}
