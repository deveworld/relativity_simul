package physics

import (
	"math"
	"testing"
)

// TestParticleCreation tests the creation of a new Particle
func TestParticleCreation(t *testing.T) {
	// RED PHASE: This test will fail because Particle doesn't exist yet
	p := NewParticle(1.0, 10.0, 20.0, 30.0, 0.1, 0.2, 0.3)

	if p.Mass != 1.0 {
		t.Errorf("Expected mass 1.0, got %f", p.Mass)
	}

	if p.Position.X != 10.0 || p.Position.Y != 20.0 || p.Position.Z != 30.0 {
		t.Errorf("Expected position (10, 20, 30), got (%f, %f, %f)",
			p.Position.X, p.Position.Y, p.Position.Z)
	}

	if p.Velocity.X != 0.1 || p.Velocity.Y != 0.2 || p.Velocity.Z != 0.3 {
		t.Errorf("Expected velocity (0.1, 0.2, 0.3), got (%f, %f, %f)",
			p.Velocity.X, p.Velocity.Y, p.Velocity.Z)
	}
}

// TestParticleRadius tests particle radius calculation based on mass
func TestParticleRadius(t *testing.T) {
	tests := []struct {
		mass           float64
		expectedRadius float64
	}{
		{1.0, 0.01},
		{10.0, 0.02},
		{100.0, 0.03},
	}

	for _, test := range tests {
		p := NewParticle(test.mass, 0, 0, 0, 0, 0, 0)
		expectedRadius := float32(math.Pow(test.mass, 1.0/3.0) * 0.01)

		if math.Abs(float64(p.Radius-expectedRadius)) > 0.001 {
			t.Errorf("For mass %f, expected radius %f, got %f",
				test.mass, expectedRadius, p.Radius)
		}
	}
}

// TestParticleKineticEnergy tests kinetic energy calculation
func TestParticleKineticEnergy(t *testing.T) {
	p := NewParticle(2.0, 0, 0, 0, 3.0, 4.0, 0) // velocity magnitude = 5.0

	ke := p.KineticEnergy()
	expected := 0.5 * 2.0 * 25.0 // 0.5 * mass * v^2

	if math.Abs(float64(ke)-expected) > 0.001 {
		t.Errorf("Expected kinetic energy %f, got %f", expected, ke)
	}
}
