/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "math/rand"

type LinkedListNode struct {
	Next  *LinkedListNode `json:"next"`
	Value interface{}     `json:"value"`
}

type LinkedList struct {
	Head  *LinkedListNode `json:"head"`
	Tail  *LinkedListNode `json:"tail"`
	Count int             `json:"count"`
}

func NewLinkedList() *LinkedList {
	list := &LinkedList{}
	list.Head = nil
	list.Tail = nil
	list.Count = 0

	return list
}

func (list *LinkedList) Remove(value interface{}) {
	var iter *LinkedListNode = list.Head

	if list.Head.Value == value {
		list.Head = list.Head.Next
		list.Tail = list.Head

		list.Count--
		return
	}

	if list.Tail.Value == value {
		for iter.Next != list.Tail {
			iter = iter.Next
		}

		list.Tail = iter
		list.Tail.Next = nil
		list.Count--
		return
	}

	iter = list.Head
	for iter.Next.Value != value {
		iter = iter.Next

		if iter.Next == nil {
			break
		}
	}

	if iter != nil && iter.Next != nil {
		iter.Next = iter.Next.Next
		list.Count--
	}
}

func (list *LinkedList) Insert(value interface{}) {
	node := &LinkedListNode{}
	node.Value = value

	if list.Head == nil {
		list.Head = node
		list.Tail = node
		list.Count++
		return
	}

	list.Tail.Next = node
	list.Tail = list.Tail.Next
	list.Count++
}

func (list *LinkedList) GetRandomNode() *LinkedListNode {
	c := len(list.Values())
	if c <= 0 {
		return nil
	}

	choice := rand.Intn(c)

	i := 0
	for iter := list.Head; iter != nil; iter = iter.Next {
		if i == choice {
			return iter
		}

		i++
	}

	return nil
}

/* Utility method to concatenate one linked list to another */
func (list *LinkedList) Concat(other *LinkedList) *LinkedList {
	for iter := other.Head; iter != nil; iter = iter.Next {
		v := iter.Value

		list.Insert(v)
	}

	return list
}

func (list *LinkedList) Contains(value interface{}) bool {
	for iter := list.Head; iter != nil; iter = iter.Next {
		v := iter.Value

		if v == value {
			return true
		}
	}

	return false
}

func (list *LinkedList) Values() []interface{} {
	var values []interface{} = make([]interface{}, 0)

	for iter := list.Head; iter != nil; iter = iter.Next {
		values = append(values, iter.Value)
	}

	return values
}
