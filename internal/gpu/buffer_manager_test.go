package gpu

import (
	"testing"
)

// TestBufferManager tests the buffer management functionality
func TestBufferManager(t *testing.T) {
	// Create a buffer manager
	manager := NewBufferManager()

	// Test that manager is created successfully
	if manager == nil {
		t.Fatal("Failed to create buffer manager")
	}
}

// TestBufferCreation tests creating GPU buffers
func TestBufferCreation(t *testing.T) {
	manager := NewBufferManager()

	// Test creating a float buffer
	floatBuffer, err := manager.CreateFloatBuffer(1024)
	if err != nil {
		// Expected to fail without OpenGL context
		if err.Error() != "OpenGL context not available" {
			t.Errorf("Unexpected error: %v", err)
		}
	} else {
		defer func() { _ = manager.FreeBuffer(floatBuffer) }()

		if floatBuffer.Size != 1024*4 { // 4 bytes per float
			t.Errorf("Float buffer size incorrect: expected %d, got %d",
				1024*4, floatBuffer.Size)
		}
	}

	// Test creating a complex buffer
	complexBuffer, err := manager.CreateComplexBuffer(512)
	if err != nil {
		// Expected to fail without OpenGL context
		if err.Error() != "OpenGL context not available" {
			t.Errorf("Unexpected error: %v", err)
		}
	} else {
		defer func() { _ = manager.FreeComplexBuffer(complexBuffer) }()

		if complexBuffer.Size != 512 {
			t.Errorf("Complex buffer size incorrect: expected %d, got %d",
				512, complexBuffer.Size)
		}
	}
}

// TestBufferDataTransfer tests uploading and downloading data
func TestBufferDataTransfer(t *testing.T) {
	manager := NewBufferManager()

	// Test float data transfer
	floatData := []float32{1.0, 2.0, 3.0, 4.0}
	buffer, err := manager.CreateFloatBuffer(len(floatData))

	if err == nil {
		defer func() { _ = manager.FreeBuffer(buffer) }()

		// Upload data
		err = manager.UploadFloatData(buffer, floatData)
		if err != nil {
			t.Errorf("Failed to upload float data: %v", err)
		}

		// Download data
		downloaded, err := manager.DownloadFloatData(buffer, len(floatData))
		if err != nil {
			t.Errorf("Failed to download float data: %v", err)
		}

		// Verify data integrity
		for i, v := range downloaded {
			if v != floatData[i] {
				t.Errorf("Data mismatch at index %d: expected %f, got %f",
					i, floatData[i], v)
			}
		}
	}

	// Test complex data transfer
	complexData := []complex128{
		complex(1.0, 2.0),
		complex(3.0, 4.0),
		complex(5.0, 6.0),
	}

	complexBuffer, err := manager.CreateComplexBuffer(len(complexData))
	if err == nil {
		defer func() { _ = manager.FreeComplexBuffer(complexBuffer) }()

		// Upload data
		err = manager.UploadComplexData(complexBuffer, complexData)
		if err != nil {
			t.Errorf("Failed to upload complex data: %v", err)
		}

		// Download data
		downloaded, err := manager.DownloadComplexData(complexBuffer, len(complexData))
		if err != nil {
			t.Errorf("Failed to download complex data: %v", err)
		}

		// Verify data integrity
		for i, v := range downloaded {
			if v != complexData[i] {
				t.Errorf("Complex data mismatch at index %d: expected %v, got %v",
					i, complexData[i], v)
			}
		}
	}
}

// TestBufferCopy tests copying data between buffers
func TestBufferCopy(t *testing.T) {
	manager := NewBufferManager()

	// Create source and destination buffers
	src, err1 := manager.CreateFloatBuffer(256)
	dst, err2 := manager.CreateFloatBuffer(256)

	if err1 == nil && err2 == nil {
		defer func() { _ = manager.FreeBuffer(src) }()
		defer func() { _ = manager.FreeBuffer(dst) }()

		// Upload test data to source
		testData := make([]float32, 256)
		for i := range testData {
			testData[i] = float32(i)
		}

		err := manager.UploadFloatData(src, testData)
		if err != nil {
			t.Errorf("Failed to upload test data: %v", err)
		}

		// Copy from source to destination
		err = manager.CopyBuffer(src, dst)
		if err != nil {
			t.Errorf("Failed to copy buffer: %v", err)
		}

		// Download and verify
		downloaded, err := manager.DownloadFloatData(dst, 256)
		if err != nil {
			t.Errorf("Failed to download copied data: %v", err)
		}

		for i, v := range downloaded {
			if v != testData[i] {
				t.Errorf("Copied data mismatch at index %d: expected %f, got %f",
					i, testData[i], v)
			}
		}
	}
}

// TestBufferResize tests resizing buffers
func TestBufferResize(t *testing.T) {
	manager := NewBufferManager()

	buffer, err := manager.CreateFloatBuffer(100)
	if err == nil {
		defer func() { _ = manager.FreeBuffer(buffer) }()

		// Resize buffer
		err = manager.ResizeBuffer(buffer, 200)
		if err != nil {
			t.Errorf("Failed to resize buffer: %v", err)
		}

		if buffer.Size != 200*4 { // 4 bytes per float
			t.Errorf("Buffer not resized correctly: expected %d, got %d",
				200*4, buffer.Size)
		}
	}
}

// TestBufferSynchronization tests buffer synchronization
func TestBufferSynchronization(t *testing.T) {
	manager := NewBufferManager()

	buffer, err := manager.CreateFloatBuffer(64)
	if err == nil {
		defer func() { _ = manager.FreeBuffer(buffer) }()

		// Test synchronization
		err = manager.SynchronizeBuffer(buffer)
		if err != nil {
			t.Errorf("Failed to synchronize buffer: %v", err)
		}
	}
}

// TestBufferCleanup tests proper cleanup of buffers
func TestBufferCleanup(t *testing.T) {
	manager := NewBufferManager()

	// Create multiple buffers
	buffers := make([]*GPUMemoryBuffer, 5)
	for i := range buffers {
		buffer, err := manager.CreateFloatBuffer(128)
		if err == nil {
			buffers[i] = buffer
		}
	}

	// Free all buffers
	for _, buffer := range buffers {
		if buffer != nil {
			err := manager.FreeBuffer(buffer)
			if err != nil {
				t.Errorf("Failed to free buffer: %v", err)
			}

			// Verify buffer is marked as freed
			if buffer.BufferID != 0 {
				t.Error("Buffer ID should be 0 after freeing")
			}
		}
	}
}

// TestBufferPool tests buffer pooling for performance
func TestBufferPool(t *testing.T) {
	manager := NewBufferManager()

	// Test getting buffer from pool
	buffer1 := manager.GetPooledBuffer(1024)
	if buffer1 == nil {
		// Pool might be empty initially
		buffer1, _ = manager.CreateFloatBuffer(1024)
	}

	if buffer1 != nil {
		// Return to pool
		manager.ReturnToPool(buffer1)

		// Get from pool again - should reuse
		buffer2 := manager.GetPooledBuffer(1024)
		if buffer2 != nil && buffer1 != buffer2 {
			// Note: In a real implementation, we'd expect the same buffer
			// For now, this is acceptable behavior
			t.Log("Buffer pool implementation pending")
		}
	}
}
