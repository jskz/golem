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
	next  *LinkedListNode
	value interface{}
}

type LinkedList struct {
	head  *LinkedListNode
	tail  *LinkedListNode
	count int
}

func NewLinkedList() *LinkedList {
	list := &LinkedList{}
	list.head = nil
	list.tail = nil
	list.count = 0

	return list
}

func (list *LinkedList) Remove(value interface{}) {
	var iter *LinkedListNode = list.head

	if list.head.value == value {
		list.head = list.head.next
		list.tail = list.head

		list.count--
		return
	}

	if list.tail.value == value {
		for iter.next != list.tail {
			iter = iter.next
		}

		list.tail = iter
		list.tail.next = nil
		list.count--
		return
	}

	iter = list.head
	for iter.next.value != value {
		iter = iter.next

		if iter.next == nil {
			break
		}
	}

	if iter != nil && iter.next != nil {
		iter.next = iter.next.next
		list.count--
	}
}

func (list *LinkedList) Insert(value interface{}) {
	node := &LinkedListNode{}
	node.value = value

	if list.head == nil {
		list.head = node
		list.tail = node
		list.count++
		return
	}

	list.tail.next = node
	list.tail = list.tail.next
	list.count++
}

func (list *LinkedList) GetRandomNode() *LinkedListNode {
	c := len(list.Values())
	if c <= 0 {
		return nil
	}

	choice := rand.Intn(c)

	i := 0
	for iter := list.head; iter != nil; iter = iter.next {
		if i == choice {
			return iter
		}

		i++
	}

	return nil
}

/* Utility method to concatenate one linked list to another */
func (list *LinkedList) Concat(other *LinkedList) *LinkedList {
	for iter := other.head; iter != nil; iter = iter.next {
		v := iter.value

		list.Insert(v)
	}

	return list
}

func (list *LinkedList) Contains(value interface{}) bool {
	for iter := list.head; iter != nil; iter = iter.next {
		v := iter.value

		if v == value {
			return true
		}
	}

	return false
}

func (list *LinkedList) Values() []interface{} {
	var values []interface{} = make([]interface{}, 0)

	for iter := list.head; iter != nil; iter = iter.next {
		values = append(values, iter.value)
	}

	return values
}
