/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "testing"

func TestLinkedListInsert(t *testing.T) {
	list := NewLinkedList[int]()

	list.Insert(1)
	list.Insert(2)

	if list.Count != 2 {
		t.Fatalf("list count = %d, want 2", list.Count)
	}

	if list.Head == nil || list.Head.Value != 2 {
		t.Fatalf("list head = %v, want 2", list.Head)
	}

	if list.Head.Next == nil || list.Head.Next.Value != 1 {
		t.Fatalf("second list value = %v, want 1", list.Head.Next)
	}
}

func TestLinkedListRemoveHead(t *testing.T) {
	list := NewLinkedList[int]()
	list.Insert(1)
	list.Insert(2)

	if !list.Remove(2) {
		t.Fatal("Remove(2) = false, want true")
	}

	if list.Count != 1 {
		t.Fatalf("list count = %d, want 1", list.Count)
	}

	if list.Head == nil || list.Head.Value != 1 {
		t.Fatalf("list head = %v, want 1", list.Head)
	}
}

func TestLinkedListRemoveMiddle(t *testing.T) {
	list := NewLinkedList[int]()
	list.Insert(1)
	list.Insert(2)
	list.Insert(3)

	if !list.Remove(2) {
		t.Fatal("Remove(2) = false, want true")
	}

	if list.Count != 2 {
		t.Fatalf("list count = %d, want 2", list.Count)
	}

	if list.Head == nil || list.Head.Value != 3 {
		t.Fatalf("list head = %v, want 3", list.Head)
	}

	if list.Head.Next == nil || list.Head.Next.Value != 1 {
		t.Fatalf("second list value = %v, want 1", list.Head.Next)
	}
}

func TestLinkedListRemoveMissing(t *testing.T) {
	list := NewLinkedList[int]()
	list.Insert(1)
	list.Insert(2)

	if list.Remove(3) {
		t.Fatal("Remove(3) = true, want false")
	}

	if list.Count != 2 {
		t.Fatalf("list count = %d, want 2", list.Count)
	}
}

func TestLinkedListContains(t *testing.T) {
	list := NewLinkedList[int]()
	list.Insert(1)
	list.Insert(2)

	if !list.Contains(1) {
		t.Fatal("Contains(1) = false, want true")
	}

	if list.Contains(3) {
		t.Fatal("Contains(3) = true, want false")
	}
}

func TestLinkedListConcat(t *testing.T) {
	list := NewLinkedList[int]()
	other := NewLinkedList[int]()

	list.Insert(1)
	other.Insert(2)
	other.Insert(3)

	if got := list.Concat(other); got != list {
		t.Fatal("Concat returned a different list")
	}

	values := list.Values()
	expected := []int{2, 3, 1}

	if len(values) != len(expected) {
		t.Fatalf("len(Values()) = %d, want %d", len(values), len(expected))
	}

	for i, value := range values {
		if value != expected[i] {
			t.Fatalf("Values()[%d] = %d, want %d", i, value, expected[i])
		}
	}
}

func TestLinkedListValues(t *testing.T) {
	list := NewLinkedList[int]()
	list.Insert(1)
	list.Insert(2)

	values := list.Values()
	expected := []int{2, 1}

	if len(values) != len(expected) {
		t.Fatalf("len(Values()) = %d, want %d", len(values), len(expected))
	}

	for i, value := range values {
		if value != expected[i] {
			t.Fatalf("Values()[%d] = %d, want %d", i, value, expected[i])
		}
	}
}

func TestLinkedListGetRandomNode(t *testing.T) {
	list := NewLinkedList[int]()
	list.Insert(1)

	node := list.GetRandomNode()
	if node == nil {
		t.Fatal("GetRandomNode() = nil, want node")
	}

	if node.Value != 1 {
		t.Fatalf("GetRandomNode().Value = %d, want 1", node.Value)
	}
}
