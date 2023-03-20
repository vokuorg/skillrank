package skillrank

import (
	"fmt"

	"math"
)

type node struct {
	weight   float64
	outbound float64
}

// Graph holds node and edge data.
type Graph struct {
	edges map[uint32](map[uint32]float64)
	nodes map[uint32]*node
}

// NewGraph initializes and returns a new graph.
func NewGraph() *Graph {
	return &Graph{
		edges: make(map[uint32](map[uint32]float64)),
		nodes: make(map[uint32]*node),
	}
}

// Link creates a weighted edge between a source-target node pair.
func (self *Graph) Link(source, target uint32, level string) {
	// weight is equal to the level of the skill, "basic" = 1.0, "intermediate" = 2.0, "advanced" = 4.0
	weight := 0.0

	switch level {
	case "basic":
		weight = 1.0
	case "intermediate":
		weight = 2.0
	case "advanced":
		weight = 4.0
	}

	if _, ok := self.nodes[source]; ok == false {
		self.nodes[source] = &node{
			weight:   0,
			outbound: 0,
		}
	}

	// If the edge already exists, we need to subtract the old weight
	if _, ok := self.edges[source]; ok == true {
		if _, ok := self.edges[source][target]; ok == true {
			self.nodes[source].outbound -= self.edges[source][target]
		}
	}

	// Add the new weight
	self.nodes[source].outbound += weight

	if _, ok := self.nodes[target]; ok == false {
		self.nodes[target] = &node{
			weight:   0,
			outbound: 0,
		}
	}

	if _, ok := self.edges[source]; ok == false {
		self.edges[source] = map[uint32]float64{}
	}

	self.edges[source][target] = weight
}

// Rank computes the PageRank of every node in the directed graph.
// alpha is the damping factor, usually set to 0.85.
// epsilon is the convergence criteria, usually set to a tiny value.
//
// This method will run as many iterations as needed, until the graph converges.
func (self *Graph) Rank(alpha, epsilon float64, callback func(id uint32, rank float64)) {
	delta := float64(1.0)
	inverse := 1 / float64(len(self.nodes))

	// Normalize all the edge weights so that their sum amounts to 1.
	for source := range self.edges {
		if self.nodes[source].outbound > 0 {
			for target := range self.edges[source] {
				self.edges[source][target] /= self.nodes[source].outbound
			}
		}
	}

	for key := range self.nodes {
		self.nodes[key].weight = inverse
	}

	for delta > epsilon {
		leak := float64(0)
		nodes := map[uint32]float64{}

		for key, value := range self.nodes {
			nodes[key] = value.weight

			if value.outbound == 0 {
				leak += value.weight
			}

			self.nodes[key].weight = 0
		}

		leak *= alpha

		for source := range self.nodes {
			for target, weight := range self.edges[source] {
				self.nodes[target].weight += alpha * nodes[source] * weight
			}

			self.nodes[source].weight += (1-alpha)*inverse + leak*inverse
		}

		delta = 0

		for key, value := range self.nodes {
			delta += math.Abs(value.weight - nodes[key])
		}
	}

	for key, value := range self.nodes {
		callback(key, value.weight)
	}
}

// Reset clears all the current graph data.
func (self *Graph) Reset() {
	self.edges = make(map[uint32](map[uint32]float64))
	self.nodes = make(map[uint32]*node)
}

// Return a JSON string with the skill ranking
func (self *Graph) RankInJSON(alpha, epsilon float64) string {
	var ranking = ""

	self.Rank(alpha, epsilon, func(id uint32, rank float64) {
		ranking += fmt.Sprintf("\"%d\": %f, ", id, rank)
	})

	ranking = ranking[:len(ranking)-2]
	ranking = "{" + ranking + "}"

	return ranking
}