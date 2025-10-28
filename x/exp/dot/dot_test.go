package dot

import (
	"bytes"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/types"
)

func TestWrite(t *testing.T) {
	t.Run("WritesNodesAndEdges", func(t *testing.T) {
		var buf bytes.Buffer

		// Build a small graph:
		// Group::admins (no parents)
		// User::alice (parent = Group::admins)
		// User::bob (no parents)
		groupUID := types.NewEntityUID("Group", types.String("admins"))
		aliceUID := types.NewEntityUID("User", types.String("alice"))
		bobUID := types.NewEntityUID("User", types.String("bob"))

		entities := types.EntityMap{}
		entities[groupUID] = types.Entity{
			UID:     groupUID,
			Parents: types.NewEntityUIDSet(), // no parents
		}
		entities[aliceUID] = types.Entity{
			UID:     aliceUID,
			Parents: types.NewEntityUIDSet(groupUID), // parent is group
		}
		entities[bobUID] = types.Entity{
			UID:     bobUID,
			Parents: types.NewEntityUIDSet(), // no parents
		}

		if err := Write(&buf, maps.Values(entities)); err != nil {
			t.Fatalf("ToDotStr returned error: %v", err)
		}
		out := buf.String()

		// Basic prelude should be present
		if !strings.Contains(out, "strict digraph") {
			t.Fatalf("output missing digraph prelude: %q", out)
		}

		// Each entity should be present as a node with a quoted label matching the ID
		expectedGroupNode := fmt.Sprintf("\t\t%q [label=%q]\n", groupUID, groupUID.ID)
		if !strings.Contains(out, expectedGroupNode) {
			t.Errorf("expected group node line %q not found in output:\n%s", expectedGroupNode, out)
		}

		expectedAliceNode := fmt.Sprintf("\t\t%q [label=%q]\n", aliceUID, aliceUID.ID)
		if !strings.Contains(out, expectedAliceNode) {
			t.Errorf("expected alice node line %q not found in output:\n%s", expectedAliceNode, out)
		}

		expectedBobNode := fmt.Sprintf("\t\t%q [label=%q]\n", bobUID, bobUID.ID)
		if !strings.Contains(out, expectedBobNode) {
			t.Errorf("expected bob node line %q not found in output:\n%s", expectedBobNode, out)
		}

		// Edge from alice to group should be present
		expectedEdge := fmt.Sprintf("\t%q -> %q\n", aliceUID, groupUID)
		if !strings.Contains(out, expectedEdge) {
			t.Errorf("expected edge %q not found in output:\n%s", expectedEdge, out)
		}
	})

	t.Run("NoEdgesWhenNoParents", func(t *testing.T) {
		var buf bytes.Buffer

		// Two entities of different types with no parents; output must contain nodes but no edges
		uidA := types.NewEntityUID("TypeA", types.String("a1"))
		uidB := types.NewEntityUID("TypeB", types.String("b1"))

		entities := types.EntityMap{
			uidA: {UID: uidA, Parents: types.NewEntityUIDSet()},
			uidB: {UID: uidB, Parents: types.NewEntityUIDSet()},
		}

		if err := Write(&buf, maps.Values(entities)); err != nil {
			t.Fatalf("ToDotStr returned error: %v", err)
		}
		out := buf.String()

		// Ensure nodes exist
		if !strings.Contains(out, strconv.Quote(uidA.String())) {
			t.Errorf("expected node for uidA %q not present", uidA.String())
		}
		if !strings.Contains(out, strconv.Quote(uidB.String())) {
			t.Errorf("expected node for uidB %q not present", uidB.String())
		}

		// Ensure there are no edges in the graph
		if strings.Contains(out, "->") {
			t.Errorf("did not expect any edges, but found some in output:\n%s", out)
		}
	})

	t.Run("WriterFailure", func(t *testing.T) {
		// Build entities with multiple types and parents to trigger all write paths:
		// - prelude write
		// - subgraph header write (first type)
		// - node write (first type)
		// - subgraph close write (first type)
		// - subgraph header write (second type)
		// - node write (second type)
		// - subgraph close write (second type)
		// - edge write
		// - final close write
		groupUID := types.NewEntityUID("Group", types.String("admins"))
		aliceUID := types.NewEntityUID("User", types.String("alice"))
		bobUID := types.NewEntityUID("User", types.String("bob"))

		entities := types.EntityMap{
			groupUID: {UID: groupUID, Parents: types.NewEntityUIDSet()},
			aliceUID: {UID: aliceUID, Parents: types.NewEntityUIDSet(groupUID)},
			bobUID:   {UID: bobUID, Parents: types.NewEntityUIDSet()},
		}

		// Test each failure point by allowing N successful writes before failing
		testCases := []struct {
			name             string
			allowedWrites    int
			expectedErrorMsg string
		}{
			{"FailOnPrelude", 0, "write failed"},
			{"FailOnFirstSubgraphHeader", 1, "write failed"},
			{"FailOnFirstNodeWrite", 2, "write failed"},
			{"FailOnSecondNodeWrite", 3, "write failed"},
			{"FailOnFirstSubgraphClose", 4, "write failed"},
			{"FailOnSecondSubgraphHeader", 5, "write failed"},
			{"FailOnSecondTypeNodeWrite", 6, "write failed"},
			{"FailOnSecondSubgraphClose", 7, "write failed"},
			{"FailOnEdgeWrite", 8, "write failed"},
			{"FailOnFinalClose", 9, "write failed"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				failingWriter := &failAfterNWriter{allowedWrites: tc.allowedWrites}
				err := Write(failingWriter, maps.Values(entities))
				if err == nil {
					t.Fatal("expected Write to return error when writer fails, got nil")
				}
				if !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("expected error message to contain %q, got: %v", tc.expectedErrorMsg, err)
				}
			})
		}
	})
}

// failAfterNWriter is a writer that fails after N successful writes
type failAfterNWriter struct {
	allowedWrites int
	writeCount    int
}

func (f *failAfterNWriter) Write(p []byte) (n int, err error) {
	if f.writeCount >= f.allowedWrites {
		return 0, fmt.Errorf("write failed")
	}
	f.writeCount++
	return len(p), nil
}
