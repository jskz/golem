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
	Count int             `json:"count"`
}

func NewLinkedList() *LinkedList {
	list := &LinkedList{}
	list.Head = nil
	list.Count = 0

	return list
}

func (list *LinkedList) Remove(value interface{}) {
	var iter *LinkedListNode = list.Head

	if list.Count == 0 || list.Head == nil {
		return
	}

	if list.Head.Value == value {
		list.Head = list.Head.Next
		list.Count--
		return
	}

	for iter.Next != nil {
		if iter.Next.Value == value {
			iter.Next = iter.Next.Next
			list.Count--
			return
		}

		iter = iter.Next
	}
}

func (list *LinkedList) Insert(value interface{}) {
	list.Head = &LinkedListNode{Next: list.Head, Value: value}
	list.Count++
}

func (list *LinkedList) GetRandomNode() *LinkedListNode {
	choice := rand.Intn(list.Count)

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
