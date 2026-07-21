package main

import (
	"fmt"

	"github.com/antonio-cafeo/maestro/internal/doctor"
)

func main() {

	fmt.Println("Maestro AI Runtime")
	fmt.Println()

	info := doctor.CollectSystemInfo()

	fmt.Println("System")
	fmt.Println(" OS:", info.OS)
	fmt.Println(" Arch:", info.Arch)
	fmt.Println(" CPU:", info.CPU)
	fmt.Println(" Home:", info.HomeDir)
}
