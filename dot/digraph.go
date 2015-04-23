package dot

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type Graph struct {
	Directed  bool
	Attrs     map[string]string
	Nodes     []*Node
	Edges     []*Edge
	Subgraphs []*Subgraph

	nodesByName map[string]*Node
}

type Subgraph struct {
	Graph
	Name    string
	Parent  *Graph
	Cluster bool
}

type Edge struct {
	Source string
	Dest   string
	Attrs  map[string]string
}

type Node struct {
	Name  string
	Attrs map[string]string
}

func NewGraph(attrs map[string]string) *Graph {
	return &Graph{
		Attrs:       attrs,
		nodesByName: make(map[string]*Node),
	}
}

func NewEdge(src, dst string, attrs map[string]string) *Edge {
	return &Edge{
		Source: src,
		Dest:   dst,
		Attrs:  attrs,
	}
}

func NewNode(n string, attrs map[string]string) *Node {
	return &Node{
		Name:  n,
		Attrs: attrs,
	}
}

func (g *Graph) AddSubgraph(name string) *Subgraph {
	subgraph := &Subgraph{
		Graph:  *NewGraph(map[string]string{}),
		Parent: g,
		Name:   name,
	}
	g.Subgraphs = append(g.Subgraphs, subgraph)
	return subgraph
}

func (g *Graph) AddAttr(k, v string) {
	g.Attrs[k] = v
}

func (g *Graph) AddNode(n *Node) {
	g.Nodes = append(g.Nodes, n)
	g.nodesByName[n.Name] = n
}

func (g *Graph) AddEdge(e *Edge) {
	g.Edges = append(g.Edges, e)
}

func (g *Graph) AddEdgeBetween(src, dst string, attrs map[string]string) error {
	g.AddEdge(NewEdge(src, dst, attrs))

	return nil
}

func (g *Graph) GetNode(name string) (*Node, error) {
	node, ok := g.nodesByName[name]
	if !ok {
		return nil, fmt.Errorf("Could not find node: %s", name)
	}
	return node, nil
}

func (g *Graph) String() string {
	w := NewGraphWriter()

	g.DrawHeader(w)
	w.Indent()
	g.DrawBody(w)
	w.Unindent()
	g.DrawFooter(w)

	return w.String()
}

func (g *Graph) DrawHeader(w *GraphWriter) {
	if g.Directed {
		w.Printf("digraph {\n")
	} else {
		w.Printf("graph {\n")
	}
}

func (g *Graph) DrawBody(w *GraphWriter) {
	for _, as := range attrStrings(g.Attrs) {
		w.Printf("%s\n", as)
	}

	nodeStrings := make([]string, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		nodeStrings = append(nodeStrings, n.String())
	}
	sort.Strings(nodeStrings)
	for _, ns := range nodeStrings {
		w.Printf(ns)
	}

	edgeStrings := make([]string, 0, len(g.Edges))
	for _, e := range g.Edges {
		edgeStrings = append(edgeStrings, e.String())
	}
	sort.Strings(edgeStrings)
	for _, es := range edgeStrings {
		w.Printf(es)
	}

	for _, s := range g.Subgraphs {
		s.DrawHeader(w)
		w.Indent()
		s.DrawBody(w)
		w.Unindent()
		s.DrawFooter(w)
	}
}

func (g *Graph) DrawFooter(w *GraphWriter) {
	w.Printf("}\n")
}

func (e *Edge) String() string {
	var buf bytes.Buffer
	buf.WriteString(
		fmt.Sprintf(
			"%q -> %q", e.Source, e.Dest))
	writeAttrs(&buf, e.Attrs)
	buf.WriteString("\n")

	return buf.String()
}

func (s *Subgraph) DrawHeader(w *GraphWriter) {
	name := s.Name
	if s.Cluster {
		name = fmt.Sprintf("cluster_%s", name)
	}
	w.Printf("subgraph %q {\n", name)
}

func (n *Node) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%q", n.Name))
	writeAttrs(&buf, n.Attrs)
	buf.WriteString("\n")

	return buf.String()
}

func writeAttrs(buf *bytes.Buffer, attrs map[string]string) {
	if len(attrs) > 0 {
		buf.WriteString(" [")
		buf.WriteString(strings.Join(attrStrings(attrs), ", "))
		buf.WriteString("]")
	}
}

func attrStrings(attrs map[string]string) []string {
	strings := make([]string, 0, len(attrs))
	for k, v := range attrs {
		strings = append(strings, fmt.Sprintf("%s = %q", k, v))
	}
	sort.Strings(strings)
	return strings
}
