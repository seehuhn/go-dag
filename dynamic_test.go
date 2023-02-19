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
