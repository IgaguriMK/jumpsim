package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"
)

const (
	BubbleDensity = 0.002375
	FieldSize     = 1000.0
	FieldPadding  = 80.0

	DefaultMaxHop = 100000
)

func main() {
	var density float64
	flag.Float64Var(&density, "d", BubbleDensity, "System densit [LY^-3]")
	var tryCount int
	flag.IntVar(&tryCount, "n", 1, "Try count per jump range")
	var maxHop int
	flag.IntVar(&maxHop, "m", DefaultMaxHop, "Max hop hop count for single problem")

	flag.Parse()

	probCh := make(chan Problem, 32)
	resultCh := make(chan *Result, 32)

	// gen problems
	go func() {
		id := 0
		for jumpRange := 6.8; jumpRange < 75; jumpRange += 0.05 {
			for i := 0; i < tryCount; i++ {
				probCh <- Problem{
					ID:        id,
					JumpRange: jumpRange,
					Density:   density,
					MaxHop:    maxHop,
				}
				id++
			}
		}
		close(probCh)
	}()

	// launch worker
	workerNum := runtime.NumCPU() - 2
	if workerNum < 1 {
		workerNum = 1
	}
	log.Println("Worker count:", workerNum)

	var workerWG sync.WaitGroup

	for i := 0; i < workerNum; i++ {
		workerWG.Add(1)
		go func() {
			for prob := range probCh {
				result := runSim(prob)
				resultCh <- result
			}
			workerWG.Done()
		}()
	}

	go func() {
		workerWG.Wait()
		close(resultCh)
	}()

	// collect results
	fmt.Println(`ID	Succ	Because	Density	JumpRange	Count	TotalJump	Efficiency`)

	nextID := 0
	results := make([]*Result, 0)
	for r := range resultCh {
		results = append(results, r)
		sort.Slice(results, func(i, j int) bool {
			return results[i].ID < results[j].ID
		})

		for {
			if len(results) == 0 {
				break
			}
			if results[0].ID > nextID {
				break
			}

			result := results[0]
			results = results[1:]
			nextID++

			if result.Succeeded {
				fmt.Printf(
					`%d	NA	T	%.6f	%.2f	%d	%.2f	%.4f`+"\n",
					result.ID,
					result.Density,
					result.JumpRange,
					result.Count,
					result.TotalJump,
					FieldSize/result.TotalJump,
				)
			} else {
				fmt.Printf(
					`%d	%s	F	%.6f	%.2f	NA	NA	NA`+"\n",
					result.ID,
					result.Because,
					result.Density,
					result.JumpRange,
				)
			}
		}

		// 進捗ログ
		log.Printf(
			"Done: id=%d, nextID=%d, len(results)",
			r.ID,
			nextID,
			len(results),
		)
	}
}

type Problem struct {
	ID        int
	JumpRange float64
	Density   float64
	MaxHop    int
}

type Result struct {
	ID        int
	Because   string
	Density   float64
	JumpRange float64
	Succeeded bool
	Count     int
	TotalJump float64
}

func (r Result) String() string {
	return fmt.Sprintf("{id=%d, succ=%t, because=%q}", r.ID, r.Succeeded, r.Because)
}

func runSim(prob Problem) *Result {
	jumpRange := prob.JumpRange
	density := prob.Density

	fieldSize := FieldSize + FieldPadding*2

	log.Printf("Start search id=%d.\n", prob.ID)

	systems := GenSystems(fieldSize, density)

	start := &Vec3{-FieldSize / 2, 0, 0, false}
	goal := &Vec3{FieldSize / 2, 0, 0, false}

	step := &Step{
		Pos:            start,
		LeftCandidates: systems.GetWithin(start, jumpRange),
	}

	cnt := 0
	for {
		cnt++

		// 試行回数が多すぎる
		if cnt > prob.MaxHop {
			return &Result{
				ID:        prob.ID,
				Density:   density,
				JumpRange: jumpRange,
				Succeeded: false,
				Because:   "exceed_max_hop",
			}
		}

		next, ok := step.Next(systems, jumpRange, goal)
		if !ok {
			if step.Prev == nil {
				return &Result{
					ID:        prob.ID,
					Density:   density,
					JumpRange: jumpRange,
					Succeeded: false,
					Because:   "no_route",
				}
			}
			step = step.Prev
			continue
		}

		if next.Pos.Within(goal, jumpRange*jumpRange) {
			goaled := &Step{
				Pos:  goal,
				Prev: next,
			}

			return &Result{
				ID:        prob.ID,
				Density:   density,
				JumpRange: jumpRange,
				Succeeded: true,
				Count:     goaled.Count(),
				TotalJump: goaled.TotalJump(),
			}
		}

		step = next
	}
}

type Step struct {
	Pos            *Vec3
	Prev           *Step
	LeftCandidates []*Vec3
}

func (s *Step) Next(systems *Systems, dist float64, goal *Vec3) (*Step, bool) {
	var nextPos *Vec3
	for {
		if len(s.LeftCandidates) == 0 {
			return nil, false
		}

		nextPos = s.LeftCandidates[0]
		s.LeftCandidates = s.LeftCandidates[1:]

		if !nextPos.Visited {
			break
		}
	}

	nextCandidates := systems.GetWithin(nextPos, dist)
	sort.Slice(nextCandidates, func(i, j int) bool {
		return nextCandidates[i].Dist2(goal) < nextCandidates[j].Dist2(goal)
	})

	nextPos.Visited = true
	return &Step{
		Pos:            nextPos,
		Prev:           s,
		LeftCandidates: nextCandidates,
	}, true
}

func (s *Step) Count() int {
	if s.Prev == nil {
		return 0
	}

	return 1 + s.Prev.Count()
}

func (s *Step) TotalJump() float64 {
	if s.Prev == nil {
		return 0.0
	}

	return s.Prev.Pos.Dist(s.Pos) + s.Prev.TotalJump()
}

type Systems struct {
	systems []*Vec3
}

func GenSystems(size, sysPerCube float64) *Systems {
	volume := size * size * size
	systemCount := int(volume * sysPerCube)

	systems := make([]*Vec3, 0, systemCount)

	for i := 0; i < systemCount; i++ {
		systems = append(systems, &Vec3{randVal(size), randVal(size), randVal(size), false})
	}

	sort.Slice(systems, func(i, j int) bool {
		return systems[i].X < systems[j].X
	})

	return &Systems{systems}
}

func randVal(size float64) float64 {
	return size*rand.Float64() - size/2
}

func (s *Systems) GetWithin(v *Vec3, dist float64) []*Vec3 {
	xMin := v.X - dist
	xMax := v.X + dist

	n := len(s.systems)
	minIndex := sort.Search(n, func(i int) bool { return s.systems[i].X >= xMin })
	overIndex := sort.Search(n, func(i int) bool { return s.systems[i].X >= xMax })

	results := make([]*Vec3, 0)
	dist_sq := dist * dist

	fcs := s.systems[minIndex:overIndex]
	for i := len(fcs) - 1; i >= 0; i-- {
		if fcs[i].X >= v.X-dist/2 && v.Within(fcs[i], dist_sq) && !fcs[i].Visited {
			results = append(results, fcs[i])
		}
	}

	return results
}

func (s *Systems) Size() int {
	return len(s.systems)
}

type Vec3 struct {
	X       float64
	Y       float64
	Z       float64
	Visited bool
}

func (v *Vec3) Dist(u *Vec3) float64 {
	return math.Sqrt(v.Dist2(u))
}

func (v *Vec3) Dist2(u *Vec3) float64 {
	return (v.X-u.X)*(v.X-u.X) + (v.Y-u.Y)*(v.Y-u.Y) + (v.Z-u.Z)*(v.Z-u.Z)
}

func (v *Vec3) Within(u *Vec3, dist_sq float64) bool {
	return v.Dist2(u) < dist_sq
}

func (v *Vec3) String() string {
	return fmt.Sprintf("[%.3f, %.3f, %.3f]", v.X, v.Y, v.Z)
}
