// Copyright (c) 2014-2019 Ludovic Faujet
// Licensed under the MIT license

package daemon

import (
	"testing"

	"github.com/opensourceways/mirrorbits/mirrors"
)

func TestCluster_RemoveMirrorID(t *testing.T) {
	c := &cluster{
		mirrorsIndex: []int{1, 3, 5, 7, 9},
	}
	c.RemoveMirrorID(5)
	found := false
	for _, id := range c.mirrorsIndex {
		if id == 5 {
			found = true
		}
	}
	if found {
		t.Fatalf("Expected mirror ID 5 to be removed")
	}
	if len(c.mirrorsIndex) != 4 {
		t.Fatalf("Expected 4 remaining, got %d", len(c.mirrorsIndex))
	}
}

func TestCluster_RemoveMirrorID_NotFound(t *testing.T) {
	c := &cluster{
		mirrorsIndex: []int{1, 3, 5},
	}
	c.RemoveMirrorID(99)
	if len(c.mirrorsIndex) != 3 {
		t.Fatalf("Expected 3 remaining (no removal), got %d", len(c.mirrorsIndex))
	}
}

func TestCluster_RemoveMirrorID_Empty(t *testing.T) {
	c := &cluster{
		mirrorsIndex: []int{},
	}
	c.RemoveMirrorID(1)
	if len(c.mirrorsIndex) != 0 {
		t.Fatalf("Expected 0 remaining")
	}
}

func TestCluster_AddMirror(t *testing.T) {
	c := &cluster{
		mirrorsIndex: []int{1, 3, 5},
	}
	c.AddMirror(&mirrors.Mirror{ID: 4})
	found := false
	for _, id := range c.mirrorsIndex {
		if id == 4 {
			found = true
		}
	}
	if !found {
		t.Fatalf("Expected mirror ID 4 to be added")
	}
}

func TestCluster_RemoveMirror(t *testing.T) {
	c := &cluster{
		mirrorsIndex: []int{1, 3, 5, 7},
	}
	c.RemoveMirror(&mirrors.Mirror{ID: 3})
	found := false
	for _, id := range c.mirrorsIndex {
		if id == 3 {
			found = true
		}
	}
	if found {
		t.Fatalf("Expected mirror ID 3 to be removed")
	}
}

func TestCluster_IsHandled_SingleNode(t *testing.T) {
	c := &cluster{
		nodeIndex:     0,
		nodeTotal:     1,
		mirrorsIndex:  []int{1, 2, 3, 4, 5},
	}
	if !c.IsHandled(3) {
		t.Fatalf("Expected true for single node handling all mirrors")
	}
}

func TestCluster_IsHandled_MultiNode(t *testing.T) {
	c := &cluster{
		nodeIndex:    0,
		nodeTotal:    2,
		mirrorsIndex: []int{1, 2, 3, 4},
	}
	if !c.IsHandled(1) {
		t.Fatalf("Expected node 0 to handle mirror 1")
	}
	if c.IsHandled(3) {
		t.Fatalf("Expected node 0 not to handle mirror 3")
	}
}
