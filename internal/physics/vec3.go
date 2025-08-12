package physics

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Vec3 represents a 3D vector with float64 precision
type Vec3 struct {
	X, Y, Z float64
}

// NewVec3 creates a new Vec3
func NewVec3(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

// Add returns the sum of two vectors
func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

// Sub returns the difference of two vectors
func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{
		X: v.X - other.X,
		Y: v.Y - other.Y,
		Z: v.Z - other.Z,
	}
}

// Scale returns the vector scaled by a scalar
func (v Vec3) Scale(s float64) Vec3 {
	return Vec3{
		X: v.X * s,
		Y: v.Y * s,
		Z: v.Z * s,
	}
}

// Length returns the magnitude of the vector
func (v Vec3) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Normalize returns a unit vector in the same direction
func (v Vec3) Normalize() Vec3 {
	length := v.Length()
	if length == 0 {
		return Vec3{} // Return zero vector if length is 0
	}
	return v.Scale(1.0 / length)
}

// Dot returns the dot product of two vectors
func (v Vec3) Dot(other Vec3) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Cross returns the cross product of two vectors
func (v Vec3) Cross(other Vec3) Vec3 {
	return Vec3{
		X: v.Y*other.Z - v.Z*other.Y,
		Y: v.Z*other.X - v.X*other.Z,
		Z: v.X*other.Y - v.Y*other.X,
	}
}

// ToRaylib converts Vec3 to raylib's Vector3
func (v Vec3) ToRaylib() rl.Vector3 {
	return rl.Vector3{
		X: float32(v.X),
		Y: float32(v.Y),
		Z: float32(v.Z),
	}
}

// Vec3FromRaylib converts raylib's Vector3 to Vec3
func Vec3FromRaylib(v rl.Vector3) Vec3 {
	return Vec3{
		X: float64(v.X),
		Y: float64(v.Y),
		Z: float64(v.Z),
	}
}

// NewRaylibVector3 is a helper to create a raylib Vector3 (for testing)
func NewRaylibVector3(x, y, z float32) rl.Vector3 {
	return rl.Vector3{X: x, Y: y, Z: z}
}
