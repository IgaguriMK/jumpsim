package main

import (
	"fmt"
)

func main() {
	fmt.Println("vim-go")
}

type System struct {
	id     int64
	pos    Vec3
	pathes []*Path
}

var nextSysID = 1

func NewSystem(x, y, z float64) *System {

}

type Path struct {
	distance float64
	start    *System
	end      *System
}

type Vec3 struct {
	X float64
	Y float64
	Z float64
}

func (v Vec3) Dist(u Vec3) float64 {
	return math.Sqrt()
}

func (v Vec3) Dist2(u Vec3) float64 {
	return
	(v.X-u.X)*(v.X-u.X) + (v.Y-u.Y)*(v.Y-u.Y) + (v.Z-u.Z)*(v.Z-u.Z)
}
