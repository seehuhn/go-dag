// seehuhn.de/go/dag - compute shortest paths in directed acyclic graphs
// Copyright (C) 2025  Jochen Voss <voss@seehuhn.de>
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

type vertex struct {
	pos   int
	steps int
}

func (v vertex) Before(u vertex) bool {
	if u.pos != v.pos {
		return v.pos < u.pos
	}
	return v.steps < u.steps
}

type dynamicLinearGraph struct{}

func (dynamicLinearGraph) AppendEdges(ee []int, v vertex) []int {
	return append(ee, v.pos+1)
}

func (dynamicLinearGraph) To(v vertex, e int) vertex {
	return vertex{e, v.steps + 1}
}

func (dynamicLinearGraph) Length(v vertex, e int) int {
	return 1
}

func TestDynamicLinear(t *testing.T) {
	g := dynamicLinearGraph{}
	ee, err := ShortestPathDyn[vertex, int, int](g, vertex{0, 0}, vertex{-1, 0})
	if err != ErrNoPath || ee != nil {
		t.Errorf("expected nil/ErrNoPath, got %v/%v", ee, err)
	}

	ee, err = ShortestPathDyn[vertex, int, int](g, vertex{0, 0}, vertex{0, 0})
	if err != nil || len(ee) != 0 {
		t.Errorf("expected no path, got %v/%v", ee, err)
	}

	ee, err = ShortestPathDyn[vertex, int, int](g, vertex{0, 0}, vertex{100, 0})
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

func BenchmarkDynamicLinear(b *testing.B) {
	g := dynamicLinearGraph{}
	for i := 0; i < b.N; i++ {
		ShortestPathDyn[vertex, int, int](g, vertex{0, 0}, vertex{100, 0})
	}
}

type dynamicDoubleEdgesGraph struct{}

type doubleVertex int

func (v doubleVertex) Before(u doubleVertex) bool {
	return v < u
}

func (dynamicDoubleEdgesGraph) AppendEdges(ee []bool, v doubleVertex) []bool {
	if v >= 100 {
		panic("overshot")
	}
	if v%2 == 0 {
		return append(ee, true, false)
	}
	return append(ee, false, true)
}

func (dynamicDoubleEdgesGraph) To(v doubleVertex, e bool) doubleVertex {
	return v + 1
}

func (dynamicDoubleEdgesGraph) Length(v doubleVertex, e bool) int {
	if e {
		return 1
	}
	return 2
}

func TestDynamicDoubleEdges(t *testing.T) {
	g := dynamicDoubleEdgesGraph{}
	ee, err := ShortestPathDyn[doubleVertex, bool, int](g, doubleVertex(0), doubleVertex(100))
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

func BenchmarkDynamicDoubleEdges(b *testing.B) {
	g := dynamicDoubleEdgesGraph{}
	for i := 0; i < b.N; i++ {
		ShortestPathDyn[doubleVertex, bool, int](g, doubleVertex(0), doubleVertex(100))
	}
}

// layeredGraph has width vertices in each of layers+1 layers, and an edge from
// every vertex of a layer to every vertex of the next layer.  Only vertex
// {0, 0} is reachable in layer 0, and layer `layers` is where the search stops.
//
// Unlike the graphs above, this keeps width vertices in the frontier at all
// times, so it exercises the parts of ShortestPathDyn which deal with more
// than one candidate at once.
type layeredGraph struct {
	width  int
	layers int
}

type layerVertex struct {
	layer int
	slot  int
}

func (v layerVertex) Before(u layerVertex) bool {
	if v.layer != u.layer {
		return v.layer < u.layer
	}
	return v.slot < u.slot
}

type layerEdge struct {
	slot   int
	length int
}

func (g layeredGraph) AppendEdges(ee []layerEdge, v layerVertex) []layerEdge {
	if v.layer >= g.layers {
		return ee
	}
	for s := 0; s < g.width; s++ {
		ee = append(ee, layerEdge{s, 1 + (v.slot*7+s*13+v.layer*3)%11})
	}
	return ee
}

func (g layeredGraph) To(v layerVertex, e layerEdge) layerVertex {
	return layerVertex{v.layer + 1, e.slot}
}

func (g layeredGraph) Length(v layerVertex, e layerEdge) int {
	return e.length
}

// layeredOptimum computes the length of the shortest path from {0, 0} to the
// last layer, one layer at a time, without using ShortestPathDyn.
func layeredOptimum(g layeredGraph) int {
	dist := map[layerVertex]int{{0, 0}: 0}
	var ee []layerEdge
	for l := 0; l < g.layers; l++ {
		next := make(map[layerVertex]int)
		for v, d := range dist {
			ee = g.AppendEdges(ee[:0], v)
			for _, e := range ee {
				w := g.To(v, e)
				c := d + g.Length(v, e)
				if old, ok := next[w]; !ok || c < old {
					next[w] = c
				}
			}
		}
		dist = next
	}

	best := -1
	for _, d := range dist {
		if best < 0 || d < best {
			best = d
		}
	}
	return best
}

func TestDynamicLayered(t *testing.T) {
	for _, width := range []int{1, 2, 5, 16} {
		for _, layers := range []int{1, 2, 7, 20} {
			g := layeredGraph{width: width, layers: layers}
			start := layerVertex{0, 0}
			end := layerVertex{layers, 0}

			ee, err := ShortestPathDyn[layerVertex, layerEdge, int](g, start, end)
			if err != nil {
				t.Fatalf("width %d, layers %d: %v", width, layers, err)
			}
			if len(ee) != layers {
				t.Errorf("width %d, layers %d: expected %d edges, got %d",
					width, layers, layers, len(ee))
			}

			// check that the edges really form a path, and add up the length
			total := 0
			v := start
			for _, e := range ee {
				found := false
				for _, c := range g.AppendEdges(nil, v) {
					if c == e {
						found = true
					}
				}
				if !found {
					t.Fatalf("width %d, layers %d: %v is not an edge at %v",
						width, layers, e, v)
				}
				total += g.Length(v, e)
				v = g.To(v, e)
			}
			if v.Before(end) {
				t.Errorf("width %d, layers %d: path ends at %v", width, layers, v)
			}
			if want := layeredOptimum(g); total != want {
				t.Errorf("width %d, layers %d: path length %d, want %d",
					width, layers, total, want)
			}
		}
	}
}

func benchmarkDynamicLayered(b *testing.B, width int) {
	g := layeredGraph{width: width, layers: 50}
	start := layerVertex{0, 0}
	end := layerVertex{50, 0}
	for i := 0; i < b.N; i++ {
		_, err := ShortestPathDyn[layerVertex, layerEdge, int](g, start, end)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDynamicLayered8(b *testing.B)   { benchmarkDynamicLayered(b, 8) }
func BenchmarkDynamicLayered32(b *testing.B)  { benchmarkDynamicLayered(b, 32) }
func BenchmarkDynamicLayered128(b *testing.B) { benchmarkDynamicLayered(b, 128) }
