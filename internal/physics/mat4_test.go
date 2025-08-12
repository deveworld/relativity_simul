package physics

import (
	"math"
	"testing"
)

// TestMat4Identity tests creating an identity matrix
func TestMat4Identity(t *testing.T) {
	m := Mat4Identity()

	// Check diagonal elements are 1
	if m[0][0] != 1.0 || m[1][1] != 1.0 || m[2][2] != 1.0 || m[3][3] != 1.0 {
		t.Errorf("Expected diagonal elements to be 1.0")
	}

	// Check off-diagonal elements are 0
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if i != j && m[i][j] != 0.0 {
				t.Errorf("Expected off-diagonal element [%d][%d] to be 0, got %f", i, j, m[i][j])
			}
		}
	}
}

// TestMat4Multiply tests matrix multiplication
func TestMat4Multiply(t *testing.T) {
	// Test identity multiplication
	identity := Mat4Identity()
	m := Mat4{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
		{13, 14, 15, 16},
	}

	result := m.Multiply(identity)

	// Result should be the same as m
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if math.Abs(result[i][j]-m[i][j]) > 0.001 {
				t.Errorf("Identity multiplication failed at [%d][%d]: expected %f, got %f",
					i, j, m[i][j], result[i][j])
			}
		}
	}

	// Test actual multiplication
	a := Mat4{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
		{13, 14, 15, 16},
	}
	b := Mat4{
		{2, 0, 0, 0},
		{0, 2, 0, 0},
		{0, 0, 2, 0},
		{0, 0, 0, 2},
	}

	result = a.Multiply(b)

	// Result should be a * 2
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			expected := a[i][j] * 2
			if math.Abs(result[i][j]-expected) > 0.001 {
				t.Errorf("Multiplication failed at [%d][%d]: expected %f, got %f",
					i, j, expected, result[i][j])
			}
		}
	}
}

// TestMat4Translation tests creating a translation matrix
func TestMat4Translation(t *testing.T) {
	m := Mat4Translation(10, 20, 30)

	// Check translation components
	if m[0][3] != 10 || m[1][3] != 20 || m[2][3] != 30 {
		t.Errorf("Expected translation (10, 20, 30), got (%f, %f, %f)",
			m[0][3], m[1][3], m[2][3])
	}

	// Check diagonal is identity
	if m[0][0] != 1 || m[1][1] != 1 || m[2][2] != 1 || m[3][3] != 1 {
		t.Errorf("Expected diagonal to be identity")
	}
}

// TestMat4Scale tests creating a scale matrix
func TestMat4Scale(t *testing.T) {
	m := Mat4Scale(2, 3, 4)

	// Check scale components
	if m[0][0] != 2 || m[1][1] != 3 || m[2][2] != 4 {
		t.Errorf("Expected scale (2, 3, 4), got (%f, %f, %f)",
			m[0][0], m[1][1], m[2][2])
	}

	// Check w component is 1
	if m[3][3] != 1 {
		t.Errorf("Expected w component to be 1, got %f", m[3][3])
	}
}

// TestMat4RotationY tests creating a Y-axis rotation matrix
func TestMat4RotationY(t *testing.T) {
	// Test 90 degree rotation
	angle := math.Pi / 2
	m := Mat4RotationY(angle)

	// Check cosine and sine values
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	if math.Abs(m[0][0]-cos) > 0.001 {
		t.Errorf("Expected m[0][0] = cos(%f) = %f, got %f", angle, cos, m[0][0])
	}
	if math.Abs(m[0][2]-sin) > 0.001 {
		t.Errorf("Expected m[0][2] = sin(%f) = %f, got %f", angle, sin, m[0][2])
	}
	if math.Abs(m[2][0]-(-sin)) > 0.001 {
		t.Errorf("Expected m[2][0] = -sin(%f) = %f, got %f", angle, -sin, m[2][0])
	}
	if math.Abs(m[2][2]-cos) > 0.001 {
		t.Errorf("Expected m[2][2] = cos(%f) = %f, got %f", angle, cos, m[2][2])
	}
}

// TestMat4TransformPoint tests transforming a point by a matrix
func TestMat4TransformPoint(t *testing.T) {
	// Test translation
	translation := Mat4Translation(10, 20, 30)
	point := NewVec3(1, 2, 3)

	result := translation.TransformPoint(point)

	if result.X != 11 || result.Y != 22 || result.Z != 33 {
		t.Errorf("Expected translated point (11, 22, 33), got (%f, %f, %f)",
			result.X, result.Y, result.Z)
	}

	// Test scaling
	scale := Mat4Scale(2, 3, 4)
	point = NewVec3(1, 1, 1)

	result = scale.TransformPoint(point)

	if result.X != 2 || result.Y != 3 || result.Z != 4 {
		t.Errorf("Expected scaled point (2, 3, 4), got (%f, %f, %f)",
			result.X, result.Y, result.Z)
	}
}

// TestMat4TransformVector tests transforming a vector by a matrix
func TestMat4TransformVector(t *testing.T) {
	// Vectors should not be affected by translation
	translation := Mat4Translation(10, 20, 30)
	vector := NewVec3(1, 0, 0)

	result := translation.TransformVector(vector)

	if result.X != 1 || result.Y != 0 || result.Z != 0 {
		t.Errorf("Expected vector unchanged by translation (1, 0, 0), got (%f, %f, %f)",
			result.X, result.Y, result.Z)
	}

	// Test rotation of unit vector
	rotation := Mat4RotationY(math.Pi / 2) // 90 degrees
	vector = NewVec3(1, 0, 0)

	result = rotation.TransformVector(vector)

	// After 90 degree Y rotation, (1,0,0) becomes (0,0,-1)
	if math.Abs(result.X-0) > 0.001 || math.Abs(result.Y-0) > 0.001 || math.Abs(result.Z-(-1)) > 0.001 {
		t.Errorf("Expected rotated vector (0, 0, -1), got (%f, %f, %f)",
			result.X, result.Y, result.Z)
	}
}

// TestMat4Transpose tests matrix transposition
func TestMat4Transpose(t *testing.T) {
	m := Mat4{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
		{13, 14, 15, 16},
	}

	transposed := m.Transpose()

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if transposed[i][j] != m[j][i] {
				t.Errorf("Transpose failed at [%d][%d]: expected %f, got %f",
					i, j, m[j][i], transposed[i][j])
			}
		}
	}
}
