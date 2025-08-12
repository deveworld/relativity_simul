package physics

import "math"

// Mat4 represents a 4x4 matrix for 3D transformations
type Mat4 [4][4]float64

// Mat4Identity creates a 4x4 identity matrix
func Mat4Identity() Mat4 {
	return Mat4{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

// Multiply performs matrix multiplication (this * other)
func (m Mat4) Multiply(other Mat4) Mat4 {
	var result Mat4

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			sum := 0.0
			for k := 0; k < 4; k++ {
				sum += m[i][k] * other[k][j]
			}
			result[i][j] = sum
		}
	}

	return result
}

// Mat4Translation creates a translation matrix
func Mat4Translation(x, y, z float64) Mat4 {
	return Mat4{
		{1, 0, 0, x},
		{0, 1, 0, y},
		{0, 0, 1, z},
		{0, 0, 0, 1},
	}
}

// Mat4Scale creates a scale matrix
func Mat4Scale(x, y, z float64) Mat4 {
	return Mat4{
		{x, 0, 0, 0},
		{0, y, 0, 0},
		{0, 0, z, 0},
		{0, 0, 0, 1},
	}
}

// Mat4RotationY creates a rotation matrix around the Y axis
func Mat4RotationY(angle float64) Mat4 {
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	return Mat4{
		{cos, 0, sin, 0},
		{0, 1, 0, 0},
		{-sin, 0, cos, 0},
		{0, 0, 0, 1},
	}
}

// Mat4RotationX creates a rotation matrix around the X axis
func Mat4RotationX(angle float64) Mat4 {
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	return Mat4{
		{1, 0, 0, 0},
		{0, cos, -sin, 0},
		{0, sin, cos, 0},
		{0, 0, 0, 1},
	}
}

// Mat4RotationZ creates a rotation matrix around the Z axis
func Mat4RotationZ(angle float64) Mat4 {
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	return Mat4{
		{cos, -sin, 0, 0},
		{sin, cos, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

// TransformPoint transforms a point by the matrix (includes translation)
func (m Mat4) TransformPoint(p Vec3) Vec3 {
	// Transform as a point (w = 1)
	x := m[0][0]*p.X + m[0][1]*p.Y + m[0][2]*p.Z + m[0][3]
	y := m[1][0]*p.X + m[1][1]*p.Y + m[1][2]*p.Z + m[1][3]
	z := m[2][0]*p.X + m[2][1]*p.Y + m[2][2]*p.Z + m[2][3]
	w := m[3][0]*p.X + m[3][1]*p.Y + m[3][2]*p.Z + m[3][3]

	// Perspective divide if needed
	if w != 0 && w != 1 {
		x /= w
		y /= w
		z /= w
	}

	return Vec3{X: x, Y: y, Z: z}
}

// TransformVector transforms a vector by the matrix (ignores translation)
func (m Mat4) TransformVector(v Vec3) Vec3 {
	// Transform as a vector (w = 0)
	x := m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z
	y := m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z
	z := m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z

	return Vec3{X: x, Y: y, Z: z}
}

// Transpose returns the transpose of the matrix
func (m Mat4) Transpose() Mat4 {
	var result Mat4

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result[i][j] = m[j][i]
		}
	}

	return result
}

// Mat4LookAt creates a view matrix looking from eye to target
func Mat4LookAt(eye, target, up Vec3) Mat4 {
	// Calculate forward, right, and up vectors
	forward := target.Sub(eye).Normalize()
	right := forward.Cross(up).Normalize()
	newUp := right.Cross(forward)

	// Create the view matrix
	return Mat4{
		{right.X, right.Y, right.Z, -right.Dot(eye)},
		{newUp.X, newUp.Y, newUp.Z, -newUp.Dot(eye)},
		{-forward.X, -forward.Y, -forward.Z, forward.Dot(eye)},
		{0, 0, 0, 1},
	}
}

// Mat4Perspective creates a perspective projection matrix
func Mat4Perspective(fovY, aspect, near, far float64) Mat4 {
	f := 1.0 / math.Tan(fovY/2.0)

	return Mat4{
		{f / aspect, 0, 0, 0},
		{0, f, 0, 0},
		{0, 0, (far + near) / (near - far), (2 * far * near) / (near - far)},
		{0, 0, -1, 0},
	}
}

// Mat4Orthographic creates an orthographic projection matrix
func Mat4Orthographic(left, right, bottom, top, near, far float64) Mat4 {
	return Mat4{
		{2 / (right - left), 0, 0, -(right + left) / (right - left)},
		{0, 2 / (top - bottom), 0, -(top + bottom) / (top - bottom)},
		{0, 0, -2 / (far - near), -(far + near) / (far - near)},
		{0, 0, 0, 1},
	}
}
