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

type H any

// Graph represents a directed acyclic graph.
type GraphWithHistory[edge E, history H, length L] interface {
	// AppendEdges appends the outgoing edges of the given vertex to a slice.
	// All edges must lead to vertices with index strictly greater than v.
	AppendEdges(ee []edge, v int) []edge

	// Length returns the length of edge e starting at vertex v.
	// The history h is the history of the path up to vertex v.
	Length(v int, h history, e edge) length

	// To returns the endpoint of an edge e starting at vertex v.
	To(v int, e edge) int

	// UpdateHistory updates the history h with the edge e (from vertex v).
	UpdateHistory(h history, v int, e edge) history
}

// ShortestPathHist returns the shortest path from 0 to n.
func ShortestPathHist[edge E, history H, length L](g GraphWithHistory[edge, history, length], n int) ([]edge, error) {
	if n == 0 {
		return nil, nil
	} else if n < 0 {
		return nil, ErrNoPath
	}

	type vertexInfo struct {
		shortest length
		from     int
		via      edge
		hist     history
		reached  bool
	}
	info := make([]vertexInfo, n+1)
	var ee []edge
	for v := 0; v < n; v++ {
		ee = g.AppendEdges(ee[:0], v)
		for _, e := range ee {
			w := g.To(v, e)
			if w <= v || w > n {
				continue
			}
			newLength := info[v].shortest + g.Length(v, info[v].hist, e)
			if !info[w].reached || newLength < info[w].shortest {
				info[w].shortest = newLength
				info[w].from = v
				info[w].via = e
				info[w].hist = g.UpdateHistory(info[v].hist, v, e)
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
