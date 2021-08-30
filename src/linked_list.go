/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

type LinkedListNode struct {
	next  *LinkedListNode
	value interface{}
}

type LinkedList struct {
	head *LinkedListNode
	tail *LinkedListNode
}

func NewLinkedList() *LinkedList {
	list := &LinkedList{}
	list.head = nil
	list.tail = nil

	return list
}

func (list *LinkedList) Remove(value interface{}) {
	var iter *LinkedListNode = list.head

	if list.head.value == value {
		list.head = list.head.next
		list.tail = list.head
		return
	}

	if list.tail.value == value {
		for iter.next != list.tail {
			iter = iter.next
		}

		list.tail = iter
		list.tail.next = nil
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
	}
}

func (list *LinkedList) Insert(value interface{}) {
	node := &LinkedListNode{}
	node.value = value

	if list.head == nil {
		list.head = node
		list.tail = node
		return
	}

	list.tail.next = node
	list.tail = list.tail.next
}

func (list *LinkedList) Values() []interface{} {
	var values []interface{} = make([]interface{}, 0)

	for iter := list.head; iter != nil; iter = iter.next {
		values = append(values, iter.value)
	}

	return values
}
