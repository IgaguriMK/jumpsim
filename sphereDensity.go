package main

import (
	"flag"
	"fmt"
	"math"
)

func main() {
	var radius float64
	flag.Float64Var(&radius, "r", 20, "radius")

	var count float64
	flag.Float64Var(&count, "c", 50, "count")

	flag.Parse()

	vol := 4.0 / 3.0 * math.Pi * math.Pow(radius, 3)

	fmt.Println(count / vol)
}
