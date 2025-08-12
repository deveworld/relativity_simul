package physics

import (
	"math"
)

// Particle represents a single particle in the simulation
type Particle struct {
	Position Vec3
	Velocity Vec3
	Mass     float32
	Radius   float32
}

// NewParticle creates a new particle with the given properties
func NewParticle(mass, px, py, pz, vx, vy, vz float64) *Particle {
	return &Particle{
		Mass:     float32(mass),
		Position: NewVec3(px, py, pz),
		Velocity: NewVec3(vx, vy, vz),
		Radius:   float32(math.Pow(mass, 1.0/3.0) * 0.01),
	}
}

// KineticEnergy calculates the kinetic energy of the particle
func (p *Particle) KineticEnergy() float32 {
	v2 := float32(p.Velocity.X*p.Velocity.X + p.Velocity.Y*p.Velocity.Y + p.Velocity.Z*p.Velocity.Z)
	return 0.5 * p.Mass * v2
}
