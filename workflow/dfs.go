package workflow

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import "github.com/yourbasic/graph"

const (
	white = iota
	gray
	black
)

const noParent = -1

type cycle struct {
	From int
	To   int
}

type dfsData struct {
	Time     int
	Color    []int
	Prev     []int
	Discover []int
	Finish   []int
	Cycles   []cycle
}

func dfs(g graph.Iterator) dfsData {
	n := g.Order()
	d := dfsData{
		Time:     0,
		Color:    make([]int, n),
		Prev:     make([]int, n),
		Discover: make([]int, n),
		Finish:   make([]int, n),
	}
	for v := 0; v < n; v++ {
		d.Color[v] = white
		d.Prev[v] = noParent
	}
	for v := 0; v < n; v++ {
		if d.Color[v] == white {
			d.dfsVisit(g, v)
		}
	}
	return d
}

func (d *dfsData) dfsVisit(g graph.Iterator, v int) {
	d.Color[v] = gray
	d.Time++
	d.Discover[v] = d.Time
	g.Visit(v, func(w int, c int64) (skip bool) {
		if d.Color[w] == white {
			d.Prev[w] = v
			d.dfsVisit(g, w)
		} else if d.Color[w] == gray {
			d.Cycles = append(d.Cycles, cycle{From: w, To: v})
		}
		return
	})
	d.Color[v] = black
	d.Time++
	d.Finish[v] = d.Time
}
