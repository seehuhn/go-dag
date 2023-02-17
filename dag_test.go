// seehuhn.de/go/dag - compute shortest paths in directed acyclic graphs
// Copyright (C) 2023  Jochen Voss <voss@seehuhn.de>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package dag

import "testing"

type linearGraph struct{}

func (linearGraph) Edges(v int) []int {
	return []int{v + 1}
}

func (linearGraph) AppendEdges(ee []int, v int) []int {
	return append(ee, v+1)
}

func (linearGraph) To(v, e int) int {
	return e
}

func (linearGraph) Length(v, e int) int {
	return 1
}

func TestLinear(t *testing.T) {
	g := linearGraph{}

	ee, err := ShortestPath[int, int](g, -1)
	if err != ErrNoPath || ee != nil {
		t.Errorf("expected nil/ErrNoPath, got %v/%v", ee, err)
	}

	ee, err = ShortestPath[int, int](g, 0)
	if err != nil || len(ee) != 0 {
		t.Errorf("expected no path, got %v/%v", ee, err)
	}

	ee, err = ShortestPath[int, int](g, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(ee) != 100 {
		t.Errorf("expected 100 edges, got %d", len(ee))
	}
	for i, e := range ee {
		if e != i+1 {
			t.Errorf("expected edge %d to be %d, got %d", i, i+1, e)
		}
	}
}

func BenchmarkLinear(b *testing.B) {
	g := linearGraph{}
	for i := 0; i < b.N; i++ {
		ShortestPath[int, int](g, 100)
	}
}

type doubleEdgesGraph struct{}

func (doubleEdgesGraph) AppendEdges(ee []bool, v int) []bool {
	if v >= 100 {
		panic("overshot")
	}
	if v%2 == 0 {
		return append(ee, true, false)
	}
	return append(ee, false, true)
}

func (doubleEdgesGraph) To(v int, e bool) int {
	return v + 1
}

func (doubleEdgesGraph) Length(v int, e bool) int {
	if e {
		return 1
	}
	return 2
}

func TestDoubleEdges(t *testing.T) {
	g := doubleEdgesGraph{}
	ee, err := ShortestPath[bool, int](g, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(ee) != 100 {
		t.Errorf("expected 100 edges, got %d", len(ee))
	}
	for i, e := range ee {
		if !e {
			t.Errorf("expected edge %d to be true, got false", i)
		}
	}
}

func BenchmarkDoubleEdges(b *testing.B) {
	g := doubleEdgesGraph{}
	for i := 0; i < b.N; i++ {
		ShortestPath[bool, int](g, 100)
	}
}
