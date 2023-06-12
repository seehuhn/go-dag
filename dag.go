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

import "errors"

// E describes the possible edge types.
type E any

// L describes the possible types for edge lengths.
type L interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Graph represents a directed acyclic graph.
type Graph[edge E, length L] interface {
	// AppendEdges appends the outgoing edges of the given vertex to a slice.
	// All edges must lead to vertices with index strictly greater than v.
	AppendEdges(ee []edge, v int) []edge

	// Length returns the length of edge e starting at vertex v.
	Length(v int, e edge) length

	// To returns the endpoint of an edge e starting at vertex v.
	To(v int, e edge) int
}

// ShortestPath returns the shortest path from 0 to n.
// ErrNoPath is returned if there is no path from start to a vertex >= end.
// Otherwise, the path is returned as a slice of edges.
func ShortestPath[edge E, length L](g Graph[edge, length], n int) ([]edge, error) {
	if n == 0 {
		return nil, nil
	} else if n < 0 {
		return nil, ErrNoPath
	}

	type vertexInfo struct {
		shortest length
		from     int
		via      edge
		reached  bool
	}
	info := make([]vertexInfo, n+1)
	info[0].reached = true
	var ee []edge
	for v := 0; v < n; v++ {
		if !info[v].reached {
			continue
		}
		ee = g.AppendEdges(ee[:0], v)
		for _, e := range ee {
			w := g.To(v, e)
			if w <= v || w > n {
				continue
			}
			newLength := info[v].shortest + g.Length(v, e)
			if !info[w].reached || newLength < info[w].shortest {
				info[w].shortest = newLength
				info[w].from = v
				info[w].via = e
				info[w].reached = true
			}
		}
	}

	if !info[n].reached {
		return nil, ErrNoPath
	}

	steps := 0
	for v := n; v != 0; v = info[v].from {
		steps++
	}
	path := make([]edge, steps)
	for v := n; v != 0; v = info[v].from {
		steps--
		path[steps] = info[v].via
	}
	return path, nil
}

var ErrNoPath = errors.New("no path exists")
