package tinys3cli

import (
	"log"
	"os"
	"runtime"
	"strconv"
)

const (
	minWorkerCount = 1
)

// GetDefaultWorkerCount returns the recommended number of workers based on CPU count.
func GetDefaultWorkerCount() int {
	cpuCount := runtime.NumCPU()
	if cpuCount <= 0 {
		return 4
	}
	return cpuCount * 2
}

// GetMaxWorkerCount returns the maximum safe worker count (NumCPU * 10).
func GetMaxWorkerCount() int {
	cpuCount := runtime.NumCPU()
	if cpuCount <= 0 {
		return 40
	}
	return cpuCount * 10
}

// GetWorkerCountFromEnv reads the TINYS3_JOBS environment variable.
// Returns 0 if the variable is not set or invalid.
func GetWorkerCountFromEnv() int {
	val := os.Getenv("TINYS3_JOBS")
	if val == "" {
		return 0
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Warning: Invalid TINYS3_JOBS value '%s', using default", val)
		return 0
	}
	return n
}

// ValidateWorkerCount checks if the worker count is within acceptable bounds.
// Returns the clamped value and a warning message if clamping occurred.
func ValidateWorkerCount(n int) (int, string) {
	minCount := minWorkerCount
	maxCount := GetMaxWorkerCount()

	if n < minCount {
		return minCount, "Worker count %d is below minimum %d, clamping to %d"
	}
	if n > maxCount {
		return maxCount, "Worker count %d exceeds maximum %d, clamping to %d"
	}
	return n, ""
}
