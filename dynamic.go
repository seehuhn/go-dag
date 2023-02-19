package dag

import (
	"sort"
)

type Beforer[T any] interface {
	Before(T) bool
}

// Graph represents a directed acyclic graph.
type DynamicGraph[vertex Beforer[vertex], edge E, length L] interface {
	// AppendEdges appends the outgoing edges of the given vertex to a slice.
	// All edges must lead to vertices with index strictly greater than v.
	AppendEdges(ee []edge, v vertex) []edge

	// Length returns the length of edge e starting at vertex v.
	Length(v vertex, e edge) length

	// To returns the endpoint of an edge e starting at vertex v.
	To(v vertex, e edge) vertex
}

// ShortestPathDyn returns the shortest path from start to a vertex >= end.
func ShortestPathDyn[vertex Beforer[vertex], edge E, length L](g DynamicGraph[vertex, edge, length], start, end vertex) ([]edge, error) {
	if end.Before(start) {
		return nil, ErrNoPath
	}

	type vertexInfo struct {
		v        vertex
		shortest length
		from     *vertexInfo
		via      edge
	}
	var info []*vertexInfo
	info = append(info, &vertexInfo{
		v: start,
	})

	var ee []edge
	arrived := false
	var bestLength length
	for len(info) > 0 && info[0].v.Before(end) {
		current := info[0]
		copy(info, info[1:])
		info = info[:len(info)-1]

		if arrived && current.shortest >= bestLength {
			continue
		}

		v := current.v
		ee = g.AppendEdges(ee[:0], v)
		for _, e := range ee {
			w := g.To(v, e)
			if w.Before(v) {
				continue
			}
			newLength := current.shortest + g.Length(v, e)

			if !w.Before(end) {
				if !arrived || newLength < bestLength {
					bestLength = newLength
				}
				arrived = true
			}

			// use binary search to find w in info
			n := len(info)
			idx := sort.Search(n, func(i int) bool {
				return !info[i].v.Before(w)
			})
			if idx == n || w.Before(info[idx].v) { // info[idx].v != w
				info = append(info, nil)
				copy(info[idx+1:], info[idx:])
				info[idx] = &vertexInfo{
					v:        w,
					shortest: newLength,
					from:     current,
					via:      e,
				}
			} else if newLength < info[idx].shortest {
				info[idx].shortest = newLength
				info[idx].from = current
				info[idx].via = e
			}
		}
	}

	if len(info) == 0 {
		return nil, ErrNoPath
	}

	bestIdx := 0
	bestLength = info[0].shortest
	for idx := 1; idx < len(info); idx++ {
		if info[idx].shortest < bestLength {
			bestIdx = idx
			bestLength = info[idx].shortest
		}
	}

	steps := 0
	for i := info[bestIdx]; i.from != nil; i = i.from {
		steps++
	}
	path := make([]edge, steps)
	for i := info[bestIdx]; i.from != nil; i = i.from {
		steps--
		path[steps] = i.via
	}
	return path, nil
}
