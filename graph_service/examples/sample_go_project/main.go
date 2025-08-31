package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// Calculator represents a simple calculator
type Calculator struct {
	history []string
}

// NewCalculator creates a new calculator instance
func NewCalculator() *Calculator {
	return &Calculator{
		history: make([]string, 0),
	}
}

// Add performs addition and records the operation
func (c *Calculator) Add(a, b int) int {
	result := a + b
	c.recordOperation(fmt.Sprintf("%d + %d = %d", a, b, result))
	return result
}

// Multiply performs multiplication and records the operation
func (c *Calculator) Multiply(a, b int) int {
	result := a * b
	c.recordOperation(fmt.Sprintf("%d * %d = %d", a, b, result))
	return result
}

// GetHistory returns the calculation history
func (c *Calculator) GetHistory() []string {
	return c.history
}

// recordOperation stores an operation in history
func (c *Calculator) recordOperation(operation string) {
	c.history = append(c.history, operation)
	fmt.Printf("Recorded: %s\n", operation)
}

// ProcessArgs processes command line arguments and performs calculations
func ProcessArgs(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: <num1> <operation> <num2>")
	}

	num1, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid first number: %v", err)
	}

	num2, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid second number: %v", err)
	}

	calc := NewCalculator()
	var result int

	switch args[1] {
	case "+":
		result = calc.Add(num1, num2)
	case "*":
		result = calc.Multiply(num1, num2)
	default:
		return fmt.Errorf("unsupported operation: %s", args[1])
	}

	fmt.Printf("Result: %d\n", result)
	return nil
}

func main() {
	fmt.Println("Go Calculator Example")

	if len(os.Args) > 1 {
		err := ProcessArgs(os.Args[1:])
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	} else {
		// Default demo
		calc := NewCalculator()
		result1 := calc.Add(5, 3)
		result2 := calc.Multiply(result1, 2)

		fmt.Printf("Demo calculation: 5 + 3 = %d, then %d * 2 = %d\n", result1, result1, result2)

		fmt.Println("\nCalculation History:")
		for _, op := range calc.GetHistory() {
			fmt.Printf("  - %s\n", op)
		}
	}
}
