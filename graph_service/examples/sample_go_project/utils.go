package main

import (
	"fmt"
	"math"
)

// MathUtils provides utility math functions
type MathUtils struct{}

// NewMathUtils creates a new MathUtils instance
func NewMathUtils() *MathUtils {
	return &MathUtils{}
}

// Power calculates a^b
func (m *MathUtils) Power(a, b int) float64 {
	return math.Pow(float64(a), float64(b))
}

// Sqrt calculates square root
func (m *MathUtils) Sqrt(x int) float64 {
	return math.Sqrt(float64(x))
}

// FormatNumber formats a number with precision
func FormatNumber(num float64, precision int) string {
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, num)
}

// ValidateInput checks if input is valid
func ValidateInput(value int) error {
	if value < 0 {
		return fmt.Errorf("negative values not allowed: %d", value)
	}
	return nil
}
