package core_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"axlab.dev/bit/core"
)

const (
	benchSize  = 1024
	benchIter  = 1024
	debugBench = false
)

func BenchmarkBase(t *testing.B) {
	for n := 0; n < t.N; n++ {
		t0 := time.Now()
		cells := make([][2]uint64, benchSize)
		for i := range cells {
			cells[i][0] = uint64(i)
		}

		for i := 1; i < benchIter; i++ {
			cur := (i - 1) % 2
			new := i % 2
			for j := range cells {
				val := (cells[j][cur] + cells[(j+1)%benchSize][cur]) / 2
				cells[j][new] = val
			}
		}

		sum := uint64(0)
		for i := range cells {
			sum += cells[i][1]
		}
		dur := time.Since(t0)
		if debugBench {
			fmt.Println("BASE:   SUM is", sum, " -- took ", dur.String())
		}
	}
}

func BenchmarkMap(t *testing.B) {
	for n := 0; n < t.N; n++ {
		t0 := time.Now()

		cells := make([]sync.Map, benchSize)
		for i := range cells {
			cells[i].Store(uint64(1), uint64(i))
		}

		for i := 2; i <= benchIter; i++ {
			gen := uint64(i)
			for j := range cells {
				var v0, v1 uint64
				if v, ok := cells[j].Load(gen - 1); ok {
					v0 = v.(uint64)
				}
				if v, ok := cells[(j+1)%benchSize].Load(gen - 1); ok {
					v1 = v.(uint64)
				}
				val := (v0 + v1) / 2
				cells[j].Store(gen, val)
			}
		}

		sum := uint64(0)
		for i := range cells {
			v, _ := cells[i].Load(uint64(benchIter))
			sum += v.(uint64)
		}

		dur := time.Since(t0)
		if debugBench {
			fmt.Println("MAP:    SUM is", sum, " -- took ", dur.String())
		}
	}
}

func BenchmarkTable(t *testing.B) {
	for n := 0; n < t.N; n++ {
		t0 := time.Now()

		cells := make([]core.Id, benchSize)

		tables := make([]*core.Table[uint64], benchIter)
		{
			g0 := core.NewTable[uint64]()
			w0 := g0.Write()
			for i := range cells {
				cells[i] = w0.Add(uint64(i))
			}
			tables[0] = w0.Finish()
		}

		for i := 1; i < benchIter; i++ {
			cur := tables[i-1]
			new := cur.Write()
			for j := range cells {
				v0 := cur.Get(cells[j])
				v1 := cur.Get(cells[(j+1)%benchSize])
				val := (v0 + v1) / 2
				new.Set(cells[j], val)
			}
			tables[i] = new.Finish()
		}

		sum := uint64(0)
		last := tables[len(tables)-1]
		for _, id := range cells {
			sum += last.Get(id)
		}

		dur := time.Since(t0)
		if debugBench {
			fmt.Println("TABLE:  SUM is", sum, " -- took ", dur.String())
		}
	}
}

func BenchmarkDataSet(t *testing.B) {
	for n := 0; n < t.N; n++ {
		t0 := time.Now()

		tables := make([]core.DataSet[uint64], benchIter)
		for id := uint64(0); id < benchSize; id++ {
			tables[0].Set(id, id)
		}

		for i := 1; i < benchIter; i++ {
			new := tables[i-1].Clone()
			for j := uint64(0); j < benchSize; j++ {
				v0 := new.Get(j)
				v1 := new.Get((j + 1) % benchSize)
				val := (v0 + v1) / 2
				new.Set(j, val)
			}
			tables[i] = new
		}

		sum := uint64(0)
		last := tables[len(tables)-1]
		for id := uint64(0); id < benchSize; id++ {
			sum += last.Get(id)
		}

		dur := time.Since(t0)
		if debugBench {
			fmt.Println("DS:     SUM is", sum, " -- took ", dur.String())
		}
	}
}

func BenchmarkState(t *testing.B) {
	for n := 0; n < t.N; n++ {
		var state []core.State

		t0 := time.Now()

		cells := make([]core.Value[uint64], benchSize)
		state = append(state, core.NewState())

		for i := range cells {
			cells[i] = core.New[uint64]()
			cells[i].Set(state[0], uint64(i))
		}

		for i := 1; i < benchIter; i++ {
			gen := i
			state = append(state, state[gen-1].Clone())
			for j := range cells {
				val := (cells[j].Get(state[gen-1]) + cells[(j+1)%benchSize].Get(state[gen-1])) / 2
				cells[j].Set(state[gen], val)
			}
		}

		last := state[len(state)-1]
		sum := uint64(0)
		for i := range cells {
			sum += cells[i].Get(last)
		}

		dur := time.Since(t0)
		if debugBench {
			fmt.Println("STATE:  SUM is", sum, " -- took ", dur.String())
		}
	}
}
