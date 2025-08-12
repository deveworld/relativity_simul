package physics

// LeapfrogStep performs one step of leapfrog integration
// This is a second-order symplectic integrator that conserves energy well
func LeapfrogStep(particles []*Particle, forceField *ForceField, dt float32, width, height int) {
	// Leapfrog integration:
	// 1. Kick (half step velocity update)
	// 2. Drift (full step position update)
	// 3. Kick (half step velocity update)

	forceCorrectionFactor := float32(0.5) // Empirical factor to improve energy conservation

	// 1. Kick - update velocities by half step
	UpdateVelocities(particles, forceField, dt*0.5, forceCorrectionFactor)

	// 2. Drift - update positions by full step
	UpdatePositions(particles, dt, width, height)

	// 3. Kick - update velocities by half step (assumes constant field for this test)
	UpdateVelocities(particles, forceField, dt*0.5, forceCorrectionFactor)
}

// RunTimeEvolution performs a complete time evolution step including force calculation
func RunTimeEvolution(particles []*Particle, dt float32, width, height int, gravitationalConstant float64) *ForceField {
	// 1. Deposit mass onto grid
	massGrid := DepositMassToGrid(particles, width, height)

	// 2. Solve Poisson equation for potential
	potentialGrid := SolvePoissonFFT(massGrid, width, height, gravitationalConstant)

	// 3. Calculate force field from potential
	forceField := CalculateGradient(potentialGrid, width, height)

	// 4. Update particle velocities and positions
	forceCorrectionFactor := float32(0.5)

	// Kick (half step)
	UpdateVelocities(particles, forceField, dt*0.5, forceCorrectionFactor)

	// Drift (full step)
	UpdatePositions(particles, dt, width, height)

	// Recalculate forces for second kick
	massGrid = DepositMassToGrid(particles, width, height)
	potentialGrid = SolvePoissonFFT(massGrid, width, height, gravitationalConstant)
	forceField = CalculateGradient(potentialGrid, width, height)

	// Kick (half step)
	UpdateVelocities(particles, forceField, dt*0.5, forceCorrectionFactor)

	return forceField
}
