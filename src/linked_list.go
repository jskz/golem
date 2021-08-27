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
}

func NewLinkedList() *LinkedList {
	list := &LinkedList{}
	list.head = nil

	return list
}

func (list *LinkedList) Remove(value interface{}) {
	var iter *LinkedListNode = list.head
	var previous *LinkedListNode = nil

	for {
		if iter.value == value {
			if previous == nil {
				list.head = iter.next
				break
			}

			previous.next = iter.next
			break
		}

		previous = iter
		iter = iter.next
	}
}

func (list *LinkedList) Insert(value interface{}) {
	node := &LinkedListNode{}
	node.value = value

	if list.head == nil {
		list.head = node
		return
	}

	iter := list.head
	for {
		if iter.next == nil {
			iter.next = node
			break
		}

		iter = iter.next
	}
}

func (list *LinkedList) Values() []interface{} {
	var values []interface{} = make([]interface{}, 0)

	for iter := list.head; iter != nil; iter = iter.next {
		values = append(values, iter.value)
	}

	return values
}
