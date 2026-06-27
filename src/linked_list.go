/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"iter"
	"math/rand"
)

type LinkedListNode[T comparable] struct {
	Next  *LinkedListNode[T] `json:"next"`
	Value T                  `json:"value"`
}

type LinkedList[T comparable] struct {
	Head  *LinkedListNode[T] `json:"head"`
	Count int                `json:"count"`
}

func NewLinkedList[T comparable]() *LinkedList[T] {
	list := &LinkedList[T]{}
	list.Head = nil
	list.Count = 0

	return list
}

func NewAnyLinkedList() *LinkedList[interface{}] {
	return NewLinkedList[interface{}]()
}

func (list *LinkedList[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for iter := list.Head; iter != nil; iter = iter.Next {
			if !yield(iter.Value) {
				return
			}
		}
	}
}

func (list *LinkedList[T]) Remove(value T) bool {
	var iter *LinkedListNode[T] = list.Head

	if list.Count == 0 || list.Head == nil {
		return false
	}

	if list.Head.Value == value {
		list.Head = list.Head.Next
		list.Count--
		return true
	}

	for iter.Next != nil {
		if iter.Next.Value == value {
			iter.Next = iter.Next.Next
			list.Count--
			return true
		}

		iter = iter.Next
	}

	return false
}

func (list *LinkedList[T]) Insert(value T) {
	list.Head = &LinkedListNode[T]{Next: list.Head, Value: value}
	list.Count++
}

func (list *LinkedList[T]) GetRandomNode() *LinkedListNode[T] {
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
func (list *LinkedList[T]) Concat(other *LinkedList[T]) *LinkedList[T] {
	for v := range other.All() {
		list.Insert(v)
	}

	return list
}

func (list *LinkedList[T]) Contains(value T) bool {
	for v := range list.All() {
		if v == value {
			return true
		}
	}

	return false
}

func (list *LinkedList[T]) Values() []T {
	var values []T = make([]T, 0)

	for value := range list.All() {
		values = append(values, value)
	}

	return values
}
