package types_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntities(t *testing.T) {
	t.Parallel()
	t.Run("Clone", func(t *testing.T) {
		t.Parallel()
		e := types.EntityMap{
			types.EntityUID{Type: "A", ID: "A"}: {},
			types.EntityUID{Type: "A", ID: "B"}: {},
			types.EntityUID{Type: "B", ID: "A"}: {},
			types.EntityUID{Type: "B", ID: "B"}: {},
		}
		clone := e.Clone()
		testutil.Equals(t, clone, e)
	})

	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		ent := types.Entity{
			UID:        types.NewEntityUID("Type", "id"),
			Attributes: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
		}
		e := types.EntityMap{
			ent.UID: ent,
		}
		got, ok := e.Get(ent.UID)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, got, ent)
		_, ok = e.Get(types.NewEntityUID("Type", "id2"))
		testutil.Equals(t, ok, false)
	})
}

func TestEntitiesJSON(t *testing.T) {
	t.Parallel()
	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()
		e := types.EntityMap{}
		ent := types.Entity{
			UID:        types.NewEntityUID("Type", "id"),
			Parents:    types.EntityUIDSet{},
			Attributes: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
		}
		ent2 := types.Entity{
			UID:        types.NewEntityUID("Type", "id2"),
			Parents:    types.NewEntityUIDSet(ent.UID),
			Attributes: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
		}
		e[ent.UID] = ent
		e[ent2.UID] = ent2
		testutil.JSONMarshalsTo(
			t,
			e,
			`[
				{"uid": {"type": "Type", "id": "id"}, "parents": [], "attrs": {"key": 42}, "tags": {}},
				{"uid": {"type": "Type" ,"id" :"id2"}, "parents": [{"type":"Type","id":"id"}], "attrs": {"key": 42}, "tags":{}}
			]`)
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()
		b := []byte(`[{"uid":{"type":"Type","id":"id"},"parents":[],"attrs":{"key":42}}]`)
		var e types.EntityMap
		err := json.Unmarshal(b, &e)
		testutil.OK(t, err)
		want := types.EntityMap{}
		ent := types.Entity{
			UID:        types.NewEntityUID("Type", "id"),
			Parents:    types.NewEntityUIDSet(),
			Attributes: types.NewRecord(types.RecordMap{"key": types.Long(42)}),
		}
		want[ent.UID] = ent
		testutil.Equals(t, e, want)
	})

	t.Run("UnmarshalErr", func(t *testing.T) {
		t.Parallel()
		var e types.EntityMap
		err := e.UnmarshalJSON([]byte(`!@#$`))
		testutil.Error(t, err)
	})
}

func TestEntitiesToDostStr(t *testing.T) {
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

		if err := entities.ToDotStr(&buf); err != nil {
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

		if err := entities.ToDotStr(&buf); err != nil {
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
}
