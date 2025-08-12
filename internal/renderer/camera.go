package renderer

import (
	"math"
	"relativity_simulation_2d/internal/physics"
)

// ProjectionType represents the type of projection
type ProjectionType int

const (
	// ProjectionPerspective represents perspective projection
	ProjectionPerspective ProjectionType = iota
	// ProjectionOrthographic represents orthographic projection
	ProjectionOrthographic
)

// Camera represents a 3D camera
type Camera struct {
	Position physics.Vec3
	Target   physics.Vec3
	Up       physics.Vec3

	// Projection parameters
	projectionType ProjectionType
	fovY           float64 // Field of view in Y (degrees)
	aspectRatio    float64 // Width/Height
	nearPlane      float64
	farPlane       float64

	// Orthographic parameters
	left, right float64
	bottom, top float64

	// Cached matrices
	viewMatrix       physics.Mat4
	projectionMatrix physics.Mat4
	viewDirty        bool
	projectionDirty  bool
}

// NewCamera creates a new camera
func NewCamera(position, target, up physics.Vec3) *Camera {
	return &Camera{
		Position:        position,
		Target:          target,
		Up:              up,
		projectionType:  ProjectionPerspective,
		fovY:            45.0,
		aspectRatio:     16.0 / 9.0,
		nearPlane:       0.1,
		farPlane:        1000.0,
		viewDirty:       true,
		projectionDirty: true,
	}
}

// GetViewMatrix returns the view matrix
func (c *Camera) GetViewMatrix() physics.Mat4 {
	if c.viewDirty {
		c.viewMatrix = physics.Mat4LookAt(c.Position, c.Target, c.Up)
		c.viewDirty = false
	}
	return c.viewMatrix
}

// GetProjectionMatrix returns the projection matrix
func (c *Camera) GetProjectionMatrix() physics.Mat4 {
	if c.projectionDirty {
		switch c.projectionType {
		case ProjectionPerspective:
			c.projectionMatrix = physics.Mat4Perspective(
				c.fovY, c.aspectRatio, c.nearPlane, c.farPlane)
		case ProjectionOrthographic:
			c.projectionMatrix = physics.Mat4Orthographic(
				c.left, c.right, c.bottom, c.top, c.nearPlane, c.farPlane)
		}
		c.projectionDirty = false
	}
	return c.projectionMatrix
}

// SetPerspective sets perspective projection parameters
func (c *Camera) SetPerspective(fovY, aspectRatio, near, far float64) {
	c.projectionType = ProjectionPerspective
	c.fovY = fovY
	c.aspectRatio = aspectRatio
	c.nearPlane = near
	c.farPlane = far
	c.projectionDirty = true
}

// SetOrthographic sets orthographic projection parameters
func (c *Camera) SetOrthographic(left, right, bottom, top, near, far float64) {
	c.projectionType = ProjectionOrthographic
	c.left = left
	c.right = right
	c.bottom = bottom
	c.top = top
	c.nearPlane = near
	c.farPlane = far
	c.projectionDirty = true
}

// GetForward returns the forward vector
func (c *Camera) GetForward() physics.Vec3 {
	return c.Target.Sub(c.Position).Normalize()
}

// GetRight returns the right vector
func (c *Camera) GetRight() physics.Vec3 {
	forward := c.GetForward()
	return forward.Cross(c.Up).Normalize()
}

// MoveForward moves the camera forward
func (c *Camera) MoveForward(distance float64) {
	forward := c.GetForward()
	movement := forward.Scale(distance)
	c.Position = c.Position.Add(movement)
	c.Target = c.Target.Add(movement)
	c.viewDirty = true
}

// MoveRight moves the camera right
func (c *Camera) MoveRight(distance float64) {
	right := c.GetRight()
	movement := right.Scale(distance)
	c.Position = c.Position.Add(movement)
	c.Target = c.Target.Add(movement)
	c.viewDirty = true
}

// MoveUp moves the camera up
func (c *Camera) MoveUp(distance float64) {
	movement := c.Up.Scale(distance)
	c.Position = c.Position.Add(movement)
	c.Target = c.Target.Add(movement)
	c.viewDirty = true
}

// Rotate rotates the camera by yaw and pitch
func (c *Camera) Rotate(yaw, pitch float64) {
	// Calculate forward vector
	forward := c.GetForward()

	// Apply yaw rotation (around Y axis)
	cosYaw := math.Cos(yaw)
	sinYaw := math.Sin(yaw)
	newForwardX := forward.X*cosYaw - forward.Z*sinYaw
	newForwardZ := forward.X*sinYaw + forward.Z*cosYaw

	forward.X = newForwardX
	forward.Z = newForwardZ

	// Apply pitch rotation (around right axis)
	if pitch != 0 {
		// Simplified pitch - just adjust Y component
		forward.Y += pitch
		forward = forward.Normalize()
	}

	// Update target based on new forward
	c.Target = c.Position.Add(forward)
	c.viewDirty = true
}

// LookAt sets the camera to look at a target
func (c *Camera) LookAt(target physics.Vec3) {
	c.Target = target
	c.viewDirty = true
}

// IsPointInFrustum checks if a point is within the camera frustum
func (c *Camera) IsPointInFrustum(point physics.Vec3) bool {
	// Transform point to camera space
	viewMatrix := c.GetViewMatrix()
	cameraSpace := viewMatrix.TransformPoint(point)

	// Check against near and far planes
	if cameraSpace.Z > -c.nearPlane || cameraSpace.Z < -c.farPlane {
		return false
	}

	if c.projectionType == ProjectionPerspective {
		// Check against frustum sides
		halfFovY := c.fovY * math.Pi / 360.0 // Convert to radians and halve
		tanHalfFovY := math.Tan(halfFovY)
		tanHalfFovX := tanHalfFovY * c.aspectRatio

		// Project onto near plane for testing
		z := -cameraSpace.Z
		maxY := tanHalfFovY * z
		maxX := tanHalfFovX * z

		if math.Abs(cameraSpace.Y) > maxY || math.Abs(cameraSpace.X) > maxX {
			return false
		}
	} else {
		// Orthographic frustum check
		if cameraSpace.X < c.left || cameraSpace.X > c.right ||
			cameraSpace.Y < c.bottom || cameraSpace.Y > c.top {
			return false
		}
	}

	return true
}

// GetViewProjectionMatrix returns the combined view-projection matrix
func (c *Camera) GetViewProjectionMatrix() physics.Mat4 {
	view := c.GetViewMatrix()
	proj := c.GetProjectionMatrix()
	return proj.Multiply(view)
}

// SetPosition sets the camera position
func (c *Camera) SetPosition(pos physics.Vec3) {
	c.Position = pos
	c.viewDirty = true
}

// SetTarget sets the camera target
func (c *Camera) SetTarget(target physics.Vec3) {
	c.Target = target
	c.viewDirty = true
}

// SetUp sets the camera up vector
func (c *Camera) SetUp(up physics.Vec3) {
	c.Up = up
	c.viewDirty = true
}

// GetYaw returns the camera yaw angle in radians
func (c *Camera) GetYaw() float64 {
	forward := c.GetForward()
	return math.Atan2(float64(forward.Z), float64(forward.X))
}

// GetPitch returns the camera pitch angle in radians
func (c *Camera) GetPitch() float64 {
	forward := c.GetForward()
	return math.Asin(float64(forward.Y))
}
