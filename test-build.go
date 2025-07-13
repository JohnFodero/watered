// Simple test to verify the new imports compile
package main

import (
	"fmt"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Testing godotenv import...")
	_ = godotenv.Load(".env.test")
	fmt.Println("Import successful!")
}
