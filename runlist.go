package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
	"unsafe"
)

// Node is an element of a linked list
type Node struct {
	Data uint
	Next *Node
}

func main() {
	addressOrderedList := flag.Bool("m", false, "create address-ordered list")
	randomlyOrderedList := flag.Bool("r", false, "create randomly-memory-ordered list")
	countIncrement := flag.Int("i", 200000, "increment of list size")
	countBegin := flag.Int("b", 1000, "beginning list size")
	countUntil := flag.Int("u", 18000000, "sort lists up to this size")
	flag.Parse()

	rand.Seed(time.Now().UnixNano() | int64(os.Getpid()))

	if *addressOrderedList && *randomlyOrderedList {
		log.Fatal("only one of -m and -r per run")
	}

	hostname, _ := os.Hostname() // not going to fail
	fmt.Printf("# %s on %s\n", time.Now().Format(time.RFC3339), hostname)
	fmt.Printf("# Start at %d nodes, end before %d nodes, increment %d\n",
		*countBegin, *countUntil, *countIncrement)

	var listCreation func(int) *Node
	var listType string

	switch {
	case *addressOrderedList:
		listType = "memory address"
		listCreation = memoryOrderedList
	case *randomlyOrderedList:
		listType = "randomly-addressed"
		listCreation = randomAddressedList
	default:
		listType = "idomatic"
		listCreation = randomValueList
	}
	fmt.Printf("# %s list ordering\n", listType)

	fmt.Println("# list length, mean ET to walk list, overall ET for 10 walks")

	for n := *countBegin; n < *countUntil; n += *countIncrement {
		var total time.Duration
		var looping time.Duration
		var head *Node
		for i := 0; i < 10; i++ {
			// fresh, new list every iteration
			head = listCreation(n)

			listLength := 0
			before := time.Now()
			for nl := head; nl != nil; nl = nl.Next {
				listLength++
			}
			elapsed := time.Since(before)
			total += elapsed
			if listLength != n {
				log.Printf("%d list length, iteration %d, list had %d nodes, should have had %d\n",
					n, i, listLength, n,
				)
			}

			elapsed = time.Since(before)
			looping += elapsed
		}
		total /= 10.0
		fmt.Printf("%d\t%.04f\t%.04f\n", n, total.Seconds(), looping.Seconds())
	}
	fmt.Printf("# end at %s on %s\n", time.Now().Format(time.RFC3339), hostname)
}

func randomValueList(n int) *Node {

	var head *Node

	for i := 0; i < n; i++ {
		head = &Node{
			Data: uint(i),
			Next: head,
		}
	}

	return head
}

func randomAddressedList(n int) *Node {
	return recursiveMergeSort(randomValueList(n))
}

func memoryOrderedList(n int) *Node {

	head := &Node{}
	head.Data = uint(uintptr(unsafe.Pointer(head)))
	tail := head

	// Append new *Node to end of list - this will create
	// a list that has blocks of nodes in ascending address order
	for i := 1; i < n; i++ {
		nn := &Node{}
		nn.Data = uint(uintptr(unsafe.Pointer(nn)))
		tail.Next = nn
		tail = tail.Next
	}

	// sort all nodes by address, so that even blocks of nodes are
	// ordered by ascending address.
	return recursiveMergeSort(head)
}

func recursiveMergeSort(head *Node) *Node {
	if head.Next == nil {
		// single node list is sorted by definiton
		return head
	}

	// because of recursion bottoming out at a 1-long-list,
	// head points to a list of at least 2 elements.

	// Setting rabbit and turtle like this means we split an
	// odd-length-list (head) into lists of length n (right)
	// and n+1 (left).
	rabbit, turtle := head.Next, &head

	for rabbit != nil {
		turtle = &(*turtle).Next
		if rabbit = rabbit.Next; rabbit != nil {
			rabbit = rabbit.Next
		}
	}

	right := *turtle
	*turtle = nil

	left := recursiveMergeSort(head)
	right = recursiveMergeSort(right)

	// Set h, t variables so that the loop doing the merge
	// does not have to have a "if h == nil" check every iteration.
	x := &right
	if left.Data < right.Data {
		x = &left
	}

	h, t := *x, *x
	*x = (*x).Next

	// left and right are either equal in length, or right is one
	// node longer, but the "<" check might take more from one list
	// than the other. Have to check both for nil.
	for left != nil && right != nil {
		var n *Node
		if left.Data < right.Data {
			n = left
			left = left.Next
		} else {
			n = right
			right = right.Next
		}
		t.Next = n
		t = t.Next
		// At the end of this for-loop, t.Next ends up being nil
		// because of the left/right list splitting.
	}

	// Either left or right are nil. If left == nil,
	// assigning nil to t.Next is no issue.
	t.Next = left
	if right != nil {
		// but if right is nil, can't assign nil to t.Next,
		// because left was non-nil.
		t.Next = right
	}

	return h
}
