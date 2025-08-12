package gpu

import (
	"errors"
	"sync"
	"time"
)

// ComputeMode represents the compute mode
type ComputeMode int

const (
	// ModeAuto automatically selects GPU or CPU based on availability
	ModeAuto ComputeMode = iota
	// ModeCPU forces CPU computation
	ModeCPU
	// ModeGPU forces GPU computation
	ModeGPU
)

// String returns string representation of ComputeMode
func (m ComputeMode) String() string {
	switch m {
	case ModeAuto:
		return "Auto"
	case ModeCPU:
		return "CPU"
	case ModeGPU:
		return "GPU"
	default:
		return "Unknown"
	}
}

// ProcessorType represents the type of processor
type ProcessorType int

const (
	// ProcessorTypeCPU represents CPU processor
	ProcessorTypeCPU ProcessorType = iota
	// ProcessorTypeGPU represents GPU processor
	ProcessorTypeGPU
)

// Processor represents a compute processor
type Processor struct {
	Type ProcessorType
}

// GetType returns the processor type
func (p *Processor) GetType() ProcessorType {
	return p.Type
}

// GPUInfo contains information about GPU
type GPUInfo struct {
	Available bool
	Name      string
	Memory    int64
}

// PerformanceStats contains performance statistics
type PerformanceStats struct {
	CPUStats Stats
	GPUStats Stats
}

// Stats contains statistics for a processor
type Stats struct {
	Count       int
	TotalTime   float64
	AverageTime float64
}

// FallbackManager manages CPU/GPU fallback logic
type FallbackManager struct {
	mu              sync.RWMutex
	mode            ComputeMode
	gpuAvailable    bool
	lastError       error
	hasError        bool
	performanceData map[ProcessorType][]float64
}

// NewFallbackManager creates a new fallback manager
func NewFallbackManager() *FallbackManager {
	return &FallbackManager{
		mode:            ModeAuto,
		gpuAvailable:    false, // In test environment, GPU is not available
		performanceData: make(map[ProcessorType][]float64),
	}
}

// GetMode returns the current compute mode
func (m *FallbackManager) GetMode() ComputeMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

// SetMode sets the compute mode
func (m *FallbackManager) SetMode(mode ComputeMode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mode = mode
}

// IsGPUAvailable checks if GPU is available
func (m *FallbackManager) IsGPUAvailable() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.gpuAvailable
}

// GetGPUInfo returns GPU information
func (m *FallbackManager) GetGPUInfo() *GPUInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &GPUInfo{
		Available: m.gpuAvailable,
		Name:      "Mock GPU",
		Memory:    4 * 1024 * 1024 * 1024, // 4GB
	}
}

// GetProcessor returns the appropriate processor based on mode
func (m *FallbackManager) GetProcessor() *Processor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	processorType := ProcessorTypeCPU

	switch m.mode {
	case ModeGPU:
		if m.gpuAvailable && !m.hasError {
			processorType = ProcessorTypeGPU
		}
		// Fall back to CPU if GPU not available or has error
	case ModeCPU:
		processorType = ProcessorTypeCPU
	case ModeAuto:
		// Choose based on availability and performance
		if m.gpuAvailable && !m.hasError {
			// Check if GPU has better performance
			if m.isGPUFaster() {
				processorType = ProcessorTypeGPU
			}
		}
	}

	return &Processor{Type: processorType}
}

// SimulateGPUError simulates a GPU error for testing
func (m *FallbackManager) SimulateGPUError() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hasError = true
	m.lastError = errors.New("simulated GPU error")

	// Fallback to CPU
	if m.mode == ModeGPU {
		m.mode = ModeCPU
	}

	return nil
}

// HasError checks if there's an error
func (m *FallbackManager) HasError() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.hasError
}

// GetLastError returns the last error
func (m *FallbackManager) GetLastError() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastError
}

// ClearErrors clears all errors
func (m *FallbackManager) ClearErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hasError = false
	m.lastError = nil
}

// AttemptRecovery attempts to recover from GPU error
func (m *FallbackManager) AttemptRecovery() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.gpuAvailable {
		return errors.New("GPU not available")
	}

	// Simulate recovery attempt
	time.Sleep(10 * time.Millisecond) // Simulate initialization time

	// Clear error state
	m.hasError = false
	m.lastError = nil

	return nil
}

// RecordPerformance records performance metrics
func (m *FallbackManager) RecordPerformance(processorType ProcessorType, timeMs float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.performanceData[processorType] = append(m.performanceData[processorType], timeMs)
}

// GetPerformanceStats returns performance statistics
func (m *FallbackManager) GetPerformanceStats() *PerformanceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &PerformanceStats{}

	// Calculate CPU stats
	if cpuData, ok := m.performanceData[ProcessorTypeCPU]; ok && len(cpuData) > 0 {
		stats.CPUStats = m.calculateStats(cpuData)
	}

	// Calculate GPU stats
	if gpuData, ok := m.performanceData[ProcessorTypeGPU]; ok && len(gpuData) > 0 {
		stats.GPUStats = m.calculateStats(gpuData)
	}

	return stats
}

// calculateStats calculates statistics from performance data
func (m *FallbackManager) calculateStats(data []float64) Stats {
	count := len(data)
	if count == 0 {
		return Stats{}
	}

	total := 0.0
	for _, v := range data {
		total += v
	}

	return Stats{
		Count:       count,
		TotalTime:   total,
		AverageTime: total / float64(count),
	}
}

// isGPUFaster checks if GPU is faster based on recorded performance
func (m *FallbackManager) isGPUFaster() bool {
	cpuData := m.performanceData[ProcessorTypeCPU]
	gpuData := m.performanceData[ProcessorTypeGPU]

	// Not enough data
	if len(cpuData) == 0 || len(gpuData) == 0 {
		return false // Default to CPU if no data
	}

	cpuStats := m.calculateStats(cpuData)
	gpuStats := m.calculateStats(gpuData)

	return gpuStats.AverageTime < cpuStats.AverageTime
}
