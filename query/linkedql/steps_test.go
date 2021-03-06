package linkedql

import (
	"context"
	"testing"

	"github.com/cayleygraph/cayley/graph/memstore"
	"github.com/cayleygraph/quad"
	"github.com/stretchr/testify/require"
)

var singleQuadData = []quad.Quad{
	quad.MakeIRI("alice", "likes", "bob", ""),
}

var testCases = []struct {
	name    string
	data    []quad.Quad
	query   IteratorStep
	results []interface{}
}{
	{
		name:  "All Entities",
		data:  singleQuadData,
		query: &Entities{Identifiers: []EntityIdentifier{EntityIdentifierString("alice")}},
		results: []interface{}{
			quad.IRI("alice"),
		},
	},
	{
		name:  "All Vertices",
		data:  singleQuadData,
		query: &Vertex{Values: nil},
		results: []interface{}{
			quad.IRI("alice"),
			quad.IRI("likes"),
			quad.IRI("bob"),
		},
	},
	{
		name: "Select",
		data: singleQuadData,
		query: &Select{
			From: &As{
				From: &Visit{
					From: &As{
						From: &Vertex{},
						Name: "liker",
					},
					Properties: &Vertex{Values: []quad.Value{quad.IRI("likes")}},
				},
				Name: "liked",
			},
		},
		results: []interface{}{
			map[string]quad.Value{
				"liker": quad.IRI("alice"),
				"liked": quad.IRI("bob"),
			},
		},
	},
	{
		name: "Select with tags",
		data: singleQuadData,
		query: &Select{
			Tags: []string{"liker"},
			From: &As{
				From: &Visit{
					From: &As{
						From: &Vertex{},
						Name: "liker",
					},
					Properties: &Vertex{Values: []quad.Value{quad.IRI("likes")}},
				},
				Name: "liked",
			},
		},
		results: []interface{}{
			map[string]quad.Value{
				"liker": quad.IRI("alice"),
			},
		},
	},
	{
		name: "Back",
		data: singleQuadData,
		query: &Back{
			From: &Visit{
				From: &Vertex{
					Values: []quad.Value{quad.IRI("alice")},
				},
				Properties: &Vertex{
					Values: []quad.Value{
						quad.IRI("likes"),
					},
				},
			},
		},
		results: []interface{}{
			quad.IRI("alice"),
		},
	},
	{
		name: "Both",
		data: []quad.Quad{
			quad.MakeIRI("alice", "likes", "bob", ""),
			quad.MakeIRI("bob", "likes", "dan", ""),
		},
		query: &Both{
			From: &Vertex{
				Values: []quad.Value{quad.IRI("bob")},
			},
			Properties: &Vertex{Values: []quad.Value{quad.IRI("likes")}},
		},
		results: []interface{}{
			quad.IRI("alice"),
			quad.IRI("dan"),
		},
	},
	{
		name: "Count",
		data: singleQuadData,
		query: &Count{
			From: &Vertex{Values: []quad.Value{}},
		},
		results: []interface{}{
			quad.Int(4),
		},
	},
	{
		name: "Difference",
		data: singleQuadData,
		query: &Difference{
			From: &Vertex{
				Values: []quad.Value{quad.IRI("alice"), quad.IRI("likes")},
			},
			Steps: []PathStep{
				&Vertex{
					Values: []quad.Value{quad.IRI("likes")},
				},
			},
		},
		results: []interface{}{
			quad.IRI("alice"),
		},
	},
	{
		name: "Filter RegExp",
		data: []quad.Quad{
			{Subject: quad.IRI("alice"), Predicate: quad.IRI("name"), Object: quad.String("Alice"), Label: nil},
		},
		query: &Filter{
			From:   &Vertex{Values: []quad.Value{}},
			Filter: &RegExp{Pattern: "A"},
		},
		results: []interface{}{
			quad.String("Alice"),
		},
	},
	{
		name: "Filter Like",
		data: []quad.Quad{
			{Subject: quad.IRI("alice"), Predicate: quad.IRI("name"), Object: quad.String("Alice"), Label: nil},
		},
		query: &Filter{
			From:   &Vertex{Values: []quad.Value{}},
			Filter: &Like{Pattern: "a%"},
		},
		results: []interface{}{
			quad.IRI("alice"),
		},
	},
	{
		name: "Filter LessThan",
		data: []quad.Quad{
			{Subject: quad.IRI("alice"), Predicate: quad.IRI("name"), Object: quad.Int(0), Label: nil},
			{Subject: quad.IRI("alice"), Predicate: quad.IRI("name"), Object: quad.Int(1), Label: nil},
		},
		query: &LessThan{
			From:  &Vertex{Values: []quad.Value{}},
			Value: quad.Int(1),
		},
		results: []interface{}{
			quad.Int(0),
		},
	},
	{
		name: "Filter GreaterThan",
		data: []quad.Quad{
			{Subject: quad.IRI("alice"), Predicate: quad.IRI("name"), Object: quad.Int(0), Label: nil},
			{Subject: quad.IRI("alice"), Predicate: quad.IRI("name"), Object: quad.Int(1), Label: nil},
		},
		query: &GreaterThan{
			From:  &Vertex{Values: []quad.Value{}},
			Value: quad.Int(0),
		},
		results: []interface{}{
			quad.Int(1),
		},
	},
	{
		name: "Filter LessThanEquals",
		data: []quad.Quad{
			{Subject: quad.IRI("alice"), Predicate: quad.IRI("name"), Object: quad.Int(-1), Label: nil},
			{Subject: quad.IRI("alice"), Predicate: quad.IRI("name"), Object: quad.Int(0), Label: nil},
			{Subject: quad.IRI("alice"), Predicate: quad.IRI("name"), Object: quad.Int(1), Label: nil},
		},
		query: &LessThanEquals{
			From:  &Vertex{Values: []quad.Value{}},
			Value: quad.Int(0),
		},
		results: []interface{}{
			quad.Int(-1),
			quad.Int(0),
		},
	},
	{
		name: "Filter GreaterThanEquals",
		data: []quad.Quad{
			{Subject: quad.IRI("alice"), Predicate: quad.IRI("name"), Object: quad.Int(0), Label: nil},
			{Subject: quad.IRI("alice"), Predicate: quad.IRI("name"), Object: quad.Int(1), Label: nil},
			{Subject: quad.IRI("alice"), Predicate: quad.IRI("name"), Object: quad.Int(2), Label: nil},
		},
		query: &GreaterThanEquals{
			From:  &Vertex{Values: []quad.Value{}},
			Value: quad.Int(1),
		},
		results: []interface{}{
			quad.Int(1),
			quad.Int(2),
		},
	},
	{
		name: "Has",
		data: singleQuadData,
		query: &Has{
			From: &Vertex{
				Values: []quad.Value{},
			},
			Property: &Vertex{
				Values: []quad.Value{quad.IRI("likes")},
			},
			Values: []quad.Value{quad.IRI("bob")},
		},
		results: []interface{}{
			quad.IRI("alice"),
		},
	},
	{
		name: "HasReverse",
		data: singleQuadData,
		query: &HasReverse{
			From: &Vertex{
				Values: []quad.Value{},
			},
			Property: &Vertex{
				Values: []quad.Value{quad.IRI("likes")},
			},
			Values: []quad.Value{quad.IRI("alice")},
		},
		results: []interface{}{
			quad.IRI("bob"),
		},
	},
	{
		name: "ViewReverse",
		data: singleQuadData,
		query: &VisitReverse{
			From:       &Vertex{Values: []quad.Value{}},
			Properties: &Vertex{Values: []quad.Value{quad.IRI("likes")}},
		},
		results: []interface{}{
			quad.IRI("alice"),
		},
	},
	{
		name: "PropertyNames",
		data: singleQuadData,
		query: &PropertyNames{
			From: &Vertex{Values: []quad.Value{}},
		},
		results: []interface{}{
			quad.IRI("likes"),
		},
	},
	{
		name: "Intersect",
		data: []quad.Quad{
			quad.MakeIRI("bob", "likes", "alice", ""),
			quad.MakeIRI("dani", "likes", "alice", ""),
		},
		query: &Intersect{
			From: &Visit{
				From: &Vertex{Values: []quad.Value{quad.IRI("bob")}},
				Properties: &Vertex{
					Values: []quad.Value{quad.IRI("likes")},
				},
			},
			Steps: []PathStep{
				&Visit{
					From:       &Vertex{Values: []quad.Value{quad.IRI("bob")}},
					Properties: &Vertex{Values: []quad.Value{quad.IRI("likes")}},
				},
			},
		},
		results: []interface{}{
			quad.IRI("alice"),
		},
	},
	{
		name: "Is",
		data: singleQuadData,
		query: &Is{
			Values: []quad.Value{quad.IRI("bob")},
			From: &Visit{
				From: &Vertex{Values: []quad.Value{quad.IRI("alice")}},
				Properties: &Vertex{
					Values: []quad.Value{quad.IRI("likes")},
				},
			},
		},
		results: []interface{}{
			quad.IRI("bob"),
		},
	},
	{
		name: "Limit",
		data: singleQuadData,
		query: &Limit{
			Limit: 2,
			From: &Vertex{
				Values: []quad.Value{},
			},
		},
		results: []interface{}{
			quad.IRI("alice"),
			quad.IRI("likes"),
		},
	},
	{
		name: "View",
		data: singleQuadData,
		query: &Visit{
			From:       &Vertex{Values: []quad.Value{}},
			Properties: &Vertex{Values: []quad.Value{quad.IRI("likes")}},
		},
		results: []interface{}{
			quad.IRI("bob"),
		},
	},
	{
		name: "PropertyNames",
		data: singleQuadData,
		query: &PropertyNames{
			From: &Vertex{Values: []quad.Value{}},
		},
		results: []interface{}{
			quad.IRI("likes"),
		},
	},
	{
		name: "Properties",
		data: singleQuadData,
		query: &Select{
			From: &Properties{
				From:  &Vertex{Values: []quad.Value{}},
				Names: []quad.IRI{quad.IRI("likes")},
			},
		},
		results: []interface{}{
			map[string]quad.Value{
				"likes": quad.IRI("bob"),
			},
		},
	},
	{
		name: "ReversePropertyNamesAs",
		data: singleQuadData,
		query: &Select{
			From: &ReversePropertyNamesAs{
				From: &Vertex{Values: []quad.Value{}},
				Tag:  "predicate",
			},
		},
		results: []interface{}{
			map[string]quad.Value{
				"predicate": quad.IRI("likes"),
			},
		},
	},
	{
		name: "PropertyNamesAs",
		data: singleQuadData,
		query: &Select{
			From: &PropertyNamesAs{
				From: &Vertex{Values: []quad.Value{}},
				Tag:  "predicate",
			},
		},
		results: []interface{}{
			map[string]quad.Value{
				"predicate": quad.IRI("likes"),
			},
		},
	},
	{
		name: "ReverseProperties",
		data: singleQuadData,
		query: &Select{
			From: &ReverseProperties{
				From:  &Vertex{Values: []quad.Value{}},
				Names: []quad.IRI{quad.IRI("likes")},
			},
		},
		results: []interface{}{
			map[string]quad.Value{
				"likes": quad.IRI("alice"),
			},
		},
	},
	{
		name: "Skip",
		data: singleQuadData,
		query: &Skip{
			Offset: 2,
			From: &Vertex{
				Values: []quad.Value{},
			},
		},
		results: []interface{}{
			quad.IRI("bob"),
		},
	},
	{
		name: "Union",
		data: singleQuadData,
		query: &Union{
			From: &Vertex{
				Values: []quad.Value{quad.IRI("alice")},
			},
			Steps: []PathStep{
				&Vertex{
					Values: []quad.Value{quad.IRI("bob")},
				},
			},
		},
		results: []interface{}{
			quad.IRI("alice"),
			quad.IRI("bob"),
		},
	},
	{
		name: "SelectFirst",
		data: singleQuadData,
		query: &SelectFirst{
			From: &As{
				From: &Visit{
					From: &As{
						Name: "liker",
						From: &Vertex{},
					},
					Properties: &Vertex{Values: []quad.Value{quad.IRI("likes")}},
				},
				Name: "liked",
			},
		},
		results: []interface{}{
			map[string]quad.Value{
				"liker": quad.IRI("alice"),
				"liked": quad.IRI("bob"),
			},
		},
	},
	{
		name: "Unique",
		data: singleQuadData,
		query: &Unique{
			From: &Vertex{
				Values: []quad.Value{quad.IRI("alice"), quad.IRI("alice"), quad.IRI("bob")},
			},
		},
		results: []interface{}{
			quad.IRI("alice"),
			quad.IRI("bob"),
		},
	},
	{
		name: "Order",
		data: singleQuadData,
		query: &Order{
			From: &Vertex{},
		},
		results: []interface{}{
			quad.IRI("alice"),
			quad.IRI("bob"),
			quad.IRI("likes"),
		},
	},
	{
		name: "Optional",
		data: []quad.Quad{
			quad.MakeIRI("alice", "likes", "bob", ""),
			quad.MakeIRI("alice", "name", "Alice", ""),
			quad.MakeIRI("bob", "name", "Bob", ""),
		},
		query: &Select{
			From: &Optional{
				From: &Properties{
					From:  &Vertex{Values: []quad.Value{}},
					Names: []quad.IRI{quad.IRI("name")},
				},
				Step: &Properties{
					From:  &Placeholder{},
					Names: []quad.IRI{quad.IRI("likes")},
				},
			},
		},
		results: []interface{}{
			map[string]quad.Value{
				"likes": quad.IRI("bob"),
				"name":  quad.IRI("Alice"),
			},
			map[string]quad.Value{
				"name": quad.IRI("Bob"),
			},
		},
	},
	{
		name: "Where",
		data: []quad.Quad{
			quad.MakeIRI("alice", "likes", "bob", ""),
			quad.MakeIRI("alice", "name", "Alice", ""),
			quad.MakeIRI("bob", "name", "Bob", ""),
		},
		query: &Select{
			From: &As{
				From: &Where{
					From: &Vertex{},
					Steps: []PathStep{
						&As{
							From: &Visit{
								From: &Visit{
									From:       &Placeholder{},
									Properties: &Vertex{Values: []quad.Value{quad.IRI("likes")}},
								},
								Properties: &Vertex{Values: []quad.Value{quad.IRI("name")}},
							},
							Name: "likesName",
						},
						&As{
							From: &Visit{
								From:       &Placeholder{},
								Properties: &Vertex{Values: []quad.Value{quad.IRI("name")}},
							},
							Name: "name",
						},
					},
				},
				Name: "person",
			},
		},
		results: []interface{}{
			map[string]quad.Value{
				"person":    quad.IRI("alice"),
				"name":      quad.IRI("Alice"),
				"likesName": quad.IRI("Bob"),
			},
		},
	},
	{
		name: "Documents",
		data: []quad.Quad{
			quad.MakeIRI("alice", "likes", "bob", ""),
			quad.MakeIRI("alice", "name", "Alice", ""),
			quad.MakeIRI("bob", "name", "Bob", ""),
			quad.MakeIRI("bob", "likes", "alice", ""),
		},
		query: &Documents{
			From: &Properties{
				From:  &Vertex{Values: []quad.Value{}},
				Names: []quad.IRI{quad.IRI("name"), quad.IRI("likes")},
			},
		},
		results: []interface{}{
			map[string]interface{}{
				"@id":   "alice",
				"name":  []quad.Value{quad.IRI("Alice")},
				"likes": []quad.Value{quad.IRI("bob")},
			},
			map[string]interface{}{
				"@id":   "bob",
				"name":  []quad.Value{quad.IRI("Bob")},
				"likes": []quad.Value{quad.IRI("alice")},
			},
		},
	},
}

func TestLinkedQL(t *testing.T) {
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			store := memstore.New(c.data...)
			ctx := context.TODO()
			iterator, err := c.query.BuildIterator(store)
			require.NoError(t, err)
			var results []interface{}
			for iterator.Next(ctx) {
				results = append(results, iterator.Result())
			}
			require.NoError(t, iterator.Err())
			require.Equal(t, c.results, results)
		})
	}
}
