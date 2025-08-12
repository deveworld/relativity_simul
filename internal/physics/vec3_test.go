package physics

import (
	"math"
	"testing"
)

// TestVec3Creation tests creating new Vec3
func TestVec3Creation(t *testing.T) {
	v := NewVec3(1.0, 2.0, 3.0)

	if v.X != 1.0 {
		t.Errorf("Expected X=1.0, got %f", v.X)
	}
	if v.Y != 2.0 {
		t.Errorf("Expected Y=2.0, got %f", v.Y)
	}
	if v.Z != 3.0 {
		t.Errorf("Expected Z=3.0, got %f", v.Z)
	}
}

// TestVec3Add tests vector addition
func TestVec3Add(t *testing.T) {
	v1 := NewVec3(1.0, 2.0, 3.0)
	v2 := NewVec3(4.0, 5.0, 6.0)

	result := v1.Add(v2)

	if result.X != 5.0 || result.Y != 7.0 || result.Z != 9.0 {
		t.Errorf("Expected (5,7,9), got (%f,%f,%f)", result.X, result.Y, result.Z)
	}
}

// TestVec3Sub tests vector subtraction
func TestVec3Sub(t *testing.T) {
	v1 := NewVec3(5.0, 7.0, 9.0)
	v2 := NewVec3(1.0, 2.0, 3.0)

	result := v1.Sub(v2)

	if result.X != 4.0 || result.Y != 5.0 || result.Z != 6.0 {
		t.Errorf("Expected (4,5,6), got (%f,%f,%f)", result.X, result.Y, result.Z)
	}
}

// TestVec3Scale tests vector scaling
func TestVec3Scale(t *testing.T) {
	v := NewVec3(2.0, 3.0, 4.0)

	result := v.Scale(2.0)

	if result.X != 4.0 || result.Y != 6.0 || result.Z != 8.0 {
		t.Errorf("Expected (4,6,8), got (%f,%f,%f)", result.X, result.Y, result.Z)
	}
}

// TestVec3Length tests vector magnitude calculation
func TestVec3Length(t *testing.T) {
	v := NewVec3(3.0, 4.0, 0.0)

	length := v.Length()
	expected := 5.0

	if math.Abs(length-expected) > 0.001 {
		t.Errorf("Expected length %f, got %f", expected, length)
	}
}

// TestVec3Normalize tests vector normalization
func TestVec3Normalize(t *testing.T) {
	v := NewVec3(3.0, 4.0, 0.0)

	normalized := v.Normalize()
	length := normalized.Length()

	if math.Abs(length-1.0) > 0.001 {
		t.Errorf("Expected normalized length 1.0, got %f", length)
	}

	// Check direction is preserved
	expectedX := 3.0 / 5.0
	expectedY := 4.0 / 5.0

	if math.Abs(normalized.X-expectedX) > 0.001 {
		t.Errorf("Expected normalized X=%f, got %f", expectedX, normalized.X)
	}
	if math.Abs(normalized.Y-expectedY) > 0.001 {
		t.Errorf("Expected normalized Y=%f, got %f", expectedY, normalized.Y)
	}
}

// TestVec3Dot tests dot product
func TestVec3Dot(t *testing.T) {
	v1 := NewVec3(2.0, 3.0, 4.0)
	v2 := NewVec3(5.0, 6.0, 7.0)

	dot := v1.Dot(v2)
	expected := 2.0*5.0 + 3.0*6.0 + 4.0*7.0 // 10 + 18 + 28 = 56

	if math.Abs(dot-expected) > 0.001 {
		t.Errorf("Expected dot product %f, got %f", expected, dot)
	}
}

// TestVec3Cross tests cross product
func TestVec3Cross(t *testing.T) {
	v1 := NewVec3(1.0, 0.0, 0.0)
	v2 := NewVec3(0.0, 1.0, 0.0)

	cross := v1.Cross(v2)

	// i × j = k
	if cross.X != 0.0 || cross.Y != 0.0 || cross.Z != 1.0 {
		t.Errorf("Expected (0,0,1), got (%f,%f,%f)", cross.X, cross.Y, cross.Z)
	}
}

// TestVec3ToRaylib tests conversion to raylib Vector3
func TestVec3ToRaylib(t *testing.T) {
	v := NewVec3(1.5, 2.5, 3.5)

	rlVec := v.ToRaylib()

	if rlVec.X != 1.5 || rlVec.Y != 2.5 || rlVec.Z != 3.5 {
		t.Errorf("Expected raylib Vector3(1.5,2.5,3.5), got (%f,%f,%f)",
			rlVec.X, rlVec.Y, rlVec.Z)
	}
}

// TestVec3FromRaylib tests conversion from raylib Vector3
func TestVec3FromRaylib(t *testing.T) {
	rlVec := NewRaylibVector3(1.5, 2.5, 3.5)

	v := Vec3FromRaylib(rlVec)

	if v.X != 1.5 || v.Y != 2.5 || v.Z != 3.5 {
		t.Errorf("Expected Vec3(1.5,2.5,3.5), got (%f,%f,%f)",
			v.X, v.Y, v.Z)
	}
}
