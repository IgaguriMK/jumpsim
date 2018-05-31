package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
)

const (
	BubbleDensity = 0.002375
	FieldSize     = 1000.0
	FieldPadding  = 80.0
)

func main() {
	var density float64
	flag.Float64Var(&density, "d", BubbleDensity, "System density")
	var jumpRange float64
	flag.Float64Var(&jumpRange, "j", 50, "Jump range")

	flag.Parse()

	fieldSize := FieldSize + FieldPadding*2

	log.Println("Generating systems...")
	systems := GenSystems(fieldSize, density)
	log.Println("Total", systems.Size(), "systems.")

	start := Vec3{-FieldSize / 2, 0, 0}
	goal := Vec3{FieldSize / 2, 0, 0}

	step := &Step{
		Pos:            start,
		LeftCandidates: systems.GetWithin(start, jumpRange),
	}

	for {
		next, ok := step.Next(systems, jumpRange)
		if !ok {
			if step.Prev == nil {
				fmt.Println("Can't find route.")
				return
			}
			step = step.Prev
			continue
		}

		if next.Pos.Within(goal, jumpRange*jumpRange) {
			next.Print()
			fmt.Printf("Goal: %.2f\n", next.Pos.Dist(goal))
			return
		}

		log.Printf("%v -> %v\n", step.Pos, next.Pos)
		step = next
	}
}

type Step struct {
	Pos            Vec3
	Prev           *Step
	LeftCandidates []Vec3
}

func (s *Step) Next(systems *Systems, dist float64) (*Step, bool) {
	if len(s.LeftCandidates) == 0 {
		return nil, false
	}

	nextPos := s.LeftCandidates[0]
	s.LeftCandidates = s.LeftCandidates[1:]

	nextCandidates := systems.GetWithin(nextPos, dist)

	return &Step{
		Pos:            nextPos,
		Prev:           s,
		LeftCandidates: nextCandidates,
	}, true
}

func (s *Step) Print() {
	if s.Prev != nil {
		s.Prev.Print()
		fmt.Printf("%v: %.2f\n", s.Pos, s.Prev.Pos.Dist(s.Pos))
	} else {
		fmt.Printf("START: %v\n", s.Pos)
	}
}

type Systems struct {
	systems []Vec3
}

func GenSystems(size, sysPerCube float64) *Systems {
	volume := size * size * size
	systemCount := int(volume * sysPerCube)

	systems := make([]Vec3, 0, systemCount)

	for i := 0; i < systemCount; i++ {
		systems = append(systems, Vec3{randVal(size), randVal(size), randVal(size)})
	}

	sort.Slice(systems, func(i, j int) bool {
		return systems[i].X < systems[j].X
	})

	return &Systems{systems}
}

func randVal(size float64) float64 {
	return size*rand.Float64() - size/2
}

func (s *Systems) GetWithin(v Vec3, dist float64) []Vec3 {
	xMin := v.X - dist
	xMax := v.X + dist

	n := len(s.systems)
	minIndex := sort.Search(n, func(i int) bool { return s.systems[i].X >= xMin })
	overIndex := sort.Search(n, func(i int) bool { return s.systems[i].X >= xMax })

	candidates := s.systems[minIndex:overIndex]
	results := make([]Vec3, 0)
	dist_sq := dist * dist
	for i := len(candidates) - 1; i >= 0; i-- {
		if v.Within(candidates[i], dist_sq) {
			results = append(results, candidates[i])
		}
	}

	return results
}

func (s *Systems) Size() int {
	return len(s.systems)
}

type Vec3 struct {
	X float64
	Y float64
	Z float64
}

func (v Vec3) Dist(u Vec3) float64 {
	return math.Sqrt(v.Dist2(u))
}

func (v Vec3) Dist2(u Vec3) float64 {
	return (v.X-u.X)*(v.X-u.X) + (v.Y-u.Y)*(v.Y-u.Y) + (v.Z-u.Z)*(v.Z-u.Z)
}

func (v Vec3) Within(u Vec3, dist_sq float64) bool {
	return v.Dist2(u) < dist_sq
}

func (v Vec3) String() string {
	return fmt.Sprintf("[%.3f, %.3f, %.3f]", v.X, v.Y, v.Z)
}
