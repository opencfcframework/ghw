//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package topology

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jaypipes/ghw/pkg/context"
	"github.com/jaypipes/ghw/pkg/cpu"
	"github.com/jaypipes/ghw/pkg/marshal"
	"github.com/jaypipes/ghw/pkg/memory"
	"github.com/jaypipes/ghw/pkg/option"
)

// Architecture describes the overall hardware architecture. It can be either
// Symmetric Multi-Processor (SMP) or Non-Uniform Memory Access (NUMA)
type Architecture int

const (
	// SMP is a Symmetric Multi-Processor system
	ARCHITECTURE_SMP Architecture = iota
	// NUMA is a Non-Uniform Memory Access system
	ARCHITECTURE_NUMA
)

var (
	architectureString = map[Architecture]string{
		ARCHITECTURE_SMP:  "SMP",
		ARCHITECTURE_NUMA: "NUMA",
	}
)

func (a Architecture) String() string {
	return architectureString[a]
}

// NOTE(jaypipes): since serialized output is as "official" as we're going to
// get, let's lowercase the string output when serializing, in order to
// "normalize" the expected serialized output
func (a Architecture) MarshalJSON() ([]byte, error) {
	return []byte("\"" + strings.ToLower(a.String()) + "\""), nil
}

// Node is an abstract construct representing a collection of processors and
// various levels of memory cache that those processors share.  In a NUMA
// architecture, there are multiple NUMA nodes, abstracted here as multiple
// Node structs. In an SMP architecture, a single Node will be available in the
// Info struct and this single struct can be used to describe the levels of
// memory caching available to the single physical processor package's physical
// processor cores
type Node struct {
	ID        int                  `json:"id"`
	Cores     []*cpu.ProcessorCore `json:"cores"`
	Caches    []*memory.Cache      `json:"caches"`
	Distances []int                `json:"distances"`
}

func (n *Node) String() string {
	return fmt.Sprintf(
		"node #%d (%d cores)",
		n.ID,
		len(n.Cores),
	)
}

// Info describes the system topology for the host hardware
type Info struct {
	ctx          *context.Context
	Architecture Architecture `json:"architecture"`
	Nodes        []*Node      `json:"nodes"`
}

// New returns a pointer to an Info struct that contains information about the
// NUMA topology on the host system
func New(opts ...*option.Option) (*Info, error) {
	ctx := context.New(opts...)
	info := &Info{ctx: ctx}
	if err := ctx.Do(info.load); err != nil {
		return nil, err
	}
	for _, node := range info.Nodes {
		sort.Sort(memory.SortByCacheLevelTypeFirstProcessor(node.Caches))
	}
	return info, nil
}

func (i *Info) String() string {
	archStr := "SMP"
	if i.Architecture == ARCHITECTURE_NUMA {
		archStr = "NUMA"
	}
	res := fmt.Sprintf(
		"topology %s (%d nodes)",
		archStr,
		len(i.Nodes),
	)
	return res
}

// simple private struct used to encapsulate topology information in a
// top-level "topology" YAML/JSON map/object key
type topologyPrinter struct {
	Info *Info `json:"topology"`
}

// YAMLString returns a string with the topology information formatted as YAML
// under a top-level "topology:" key
func (i *Info) YAMLString() string {
	return marshal.SafeYAML(topologyPrinter{i})
}

// JSONString returns a string with the topology information formatted as JSON
// under a top-level "topology:" key
func (i *Info) JSONString(indent bool) string {
	return marshal.SafeJSON(topologyPrinter{i}, indent)
}
