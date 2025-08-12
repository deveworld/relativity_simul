package renderer

import (
	"math"
	"relativity_simulation_2d/internal/physics"
	"testing"
)

// TestCameraCreation tests creating a camera
func TestCameraCreation(t *testing.T) {
	cam := NewCamera(
		physics.NewVec3(0, 10, 20), // position
		physics.NewVec3(0, 0, 0),   // target
		physics.NewVec3(0, 1, 0),   // up
	)

	if cam == nil {
		t.Fatal("Failed to create camera")
	}

	// Check initial position
	if cam.Position.X != 0 || cam.Position.Y != 10 || cam.Position.Z != 20 {
		t.Errorf("Camera position incorrect: got %v", cam.Position)
	}

	// Check initial target
	if cam.Target.X != 0 || cam.Target.Y != 0 || cam.Target.Z != 0 {
		t.Errorf("Camera target incorrect: got %v", cam.Target)
	}
}

// TestViewMatrix tests view matrix calculation
func TestViewMatrix(t *testing.T) {
	cam := NewCamera(
		physics.NewVec3(0, 0, 10), // position
		physics.NewVec3(0, 0, 0),  // target
		physics.NewVec3(0, 1, 0),  // up
	)

	viewMatrix := cam.GetViewMatrix()

	// The view matrix should transform world to camera space
	// For a camera at (0,0,10) looking at origin, the view matrix
	// should translate points by -10 in Z
	origin := physics.NewVec3(0, 0, 0)
	transformed := viewMatrix.TransformPoint(origin)

	// Origin should be at (0, 0, -10) in camera space
	if math.Abs(transformed.Z-(-10)) > 1e-6 {
		t.Errorf("View matrix transformation incorrect: expected Z=-10, got %v", transformed.Z)
	}
}

// TestProjectionMatrix tests projection matrix calculation
func TestProjectionMatrix(t *testing.T) {
	cam := NewCamera(
		physics.NewVec3(0, 0, 10),
		physics.NewVec3(0, 0, 0),
		physics.NewVec3(0, 1, 0),
	)

	// Set projection parameters
	cam.SetPerspective(45.0, 16.0/9.0, 0.1, 100.0)

	projMatrix := cam.GetProjectionMatrix()

	// Check that the projection matrix is not identity
	identity := physics.Mat4Identity()
	if matricesEqual(projMatrix, identity) {
		t.Error("Projection matrix should not be identity")
	}

	// Test that near plane maps to -1 and far plane maps to 1 in NDC
	// This is a simplified test - full projection testing would be more complex
	nearPoint := physics.NewVec3(0, 0, -0.1)
	farPoint := physics.NewVec3(0, 0, -100)

	nearTransformed := projMatrix.TransformPoint(nearPoint)
	farTransformed := projMatrix.TransformPoint(farPoint)

	// After perspective divide, near should map close to -1, far close to 1
	// (This is a simplified check)
	if nearTransformed.Z > farTransformed.Z {
		t.Error("Projection matrix depth ordering incorrect")
	}
}

// TestCameraMovement tests camera movement operations
func TestCameraMovement(t *testing.T) {
	cam := NewCamera(
		physics.NewVec3(0, 0, 0),
		physics.NewVec3(0, 0, -1),
		physics.NewVec3(0, 1, 0),
	)

	// Test move forward
	cam.MoveForward(5.0)
	if math.Abs(cam.Position.Z-(-5.0)) > 1e-6 {
		t.Errorf("Camera forward movement incorrect: expected Z=-5, got %v", cam.Position.Z)
	}

	// Test move right
	cam.MoveRight(3.0)
	if math.Abs(cam.Position.X-3.0) > 1e-6 {
		t.Errorf("Camera right movement incorrect: expected X=3, got %v", cam.Position.X)
	}

	// Test move up
	cam.MoveUp(2.0)
	if math.Abs(cam.Position.Y-2.0) > 1e-6 {
		t.Errorf("Camera up movement incorrect: expected Y=2, got %v", cam.Position.Y)
	}
}

// TestCameraRotation tests camera rotation operations
func TestCameraRotation(t *testing.T) {
	cam := NewCamera(
		physics.NewVec3(0, 0, 0),
		physics.NewVec3(1, 0, 0), // Looking along +X
		physics.NewVec3(0, 1, 0),
	)

	// Test yaw rotation (around Y axis)
	cam.Rotate(math.Pi/2, 0) // 90 degrees yaw

	// After 90 degree yaw from +X, should be looking along +Z
	forward := cam.GetForward()
	if math.Abs(forward.Z-1.0) > 0.1 {
		t.Errorf("Camera yaw rotation incorrect: forward = %v", forward)
	}
}

// TestCameraLookAt tests the look-at functionality
func TestCameraLookAt(t *testing.T) {
	cam := NewCamera(
		physics.NewVec3(10, 10, 10),
		physics.NewVec3(0, 0, 0),
		physics.NewVec3(0, 1, 0),
	)

	// Change look-at target
	newTarget := physics.NewVec3(5, 5, 5)
	cam.LookAt(newTarget)

	// Check that target was updated
	if cam.Target != newTarget {
		t.Errorf("Camera look-at failed: expected target %v, got %v", newTarget, cam.Target)
	}

	// Check that forward vector points toward target
	forward := cam.GetForward()
	toTarget := newTarget.Sub(cam.Position).Normalize()

	dot := forward.Dot(toTarget)
	if math.Abs(dot-1.0) > 1e-6 {
		t.Errorf("Camera not looking at target: dot product = %v", dot)
	}
}

// TestOrthographicProjection tests orthographic projection
func TestOrthographicProjection(t *testing.T) {
	cam := NewCamera(
		physics.NewVec3(0, 0, 10),
		physics.NewVec3(0, 0, 0),
		physics.NewVec3(0, 1, 0),
	)

	// Set orthographic projection
	cam.SetOrthographic(-10, 10, -10, 10, 0.1, 100)

	projMatrix := cam.GetProjectionMatrix()

	// In orthographic projection, parallel lines remain parallel
	// Test that X and Y coordinates are scaled but not affected by Z
	p1 := physics.NewVec3(5, 5, -10)
	p2 := physics.NewVec3(5, 5, -50)

	t1 := projMatrix.TransformPoint(p1)
	t2 := projMatrix.TransformPoint(p2)

	// X and Y should be the same (parallel projection)
	if math.Abs(t1.X-t2.X) > 1e-6 || math.Abs(t1.Y-t2.Y) > 1e-6 {
		t.Error("Orthographic projection should preserve X,Y for different Z values")
	}
}

// TestFrustum tests frustum calculations for culling
func TestFrustum(t *testing.T) {
	cam := NewCamera(
		physics.NewVec3(0, 0, 0),
		physics.NewVec3(0, 0, -1),
		physics.NewVec3(0, 1, 0),
	)

	cam.SetPerspective(60.0, 1.0, 1.0, 100.0)

	// Test point inside frustum
	insidePoint := physics.NewVec3(0, 0, -10)
	if !cam.IsPointInFrustum(insidePoint) {
		t.Error("Point should be inside frustum")
	}

	// Test point outside frustum (behind camera)
	behindPoint := physics.NewVec3(0, 0, 10)
	if cam.IsPointInFrustum(behindPoint) {
		t.Error("Point behind camera should not be in frustum")
	}

	// Test point outside frustum (beyond far plane)
	farPoint := physics.NewVec3(0, 0, -200)
	if cam.IsPointInFrustum(farPoint) {
		t.Error("Point beyond far plane should not be in frustum")
	}
}

// Helper function to compare matrices
func matricesEqual(a, b physics.Mat4) bool {
	tolerance := 1e-10
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if math.Abs(a[i][j]-b[i][j]) > tolerance {
				return false
			}
		}
	}
	return true
}
