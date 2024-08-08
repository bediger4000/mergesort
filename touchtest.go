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

type MergeFn func(*Node, *Node) *Node

func main() {
	alreadySorted := flag.Bool("s", false, "already sorted list")
	addressOrderedList := flag.Bool("m", false, "create address-ordered list for each iteration")
	weaveNodes := flag.Bool("w", false, "weave sublists together")
	countIncrement := flag.Int("i", 200000, "increment of list size")
	countBegin := flag.Int("b", 1000, "beginning list size")
	countUntil := flag.Int("u", 18000000, "touch lists up to this size")
	flag.Parse()

	rand.Seed(time.Now().UnixNano() | int64(os.Getpid()))

	hostname, _ := os.Hostname() // not going to fail
	fmt.Printf("# %s on %s\n", time.Now().Format(time.RFC3339), hostname)
	fmt.Printf("# Start at %d nodes, end before %d nodes, increment %d\n",
		*countBegin, *countUntil, *countIncrement)
	sortType := "bottom-up iterative"
	touchType := "data incrementing"
	if *weaveNodes {
		touchType = "sub-list weaving"
	}
	fmt.Printf("# %s style list %s\n", sortType, touchType)
	listType := "idiomatic"
	if *addressOrderedList {
		listType = "memory address"
	}
	fmt.Printf("# %s list ordering\n", listType)
	fmt.Printf("# nodes %d bytes in size\n", unsafe.Sizeof(Node{}))

	nodeAccess := "increment data only"
	mergeFn := incrementData
	if *weaveNodes {
		nodeAccess = "weave sublists"
		mergeFn = weaveSublists
	}
	fmt.Printf("# node acccess is %s\n", nodeAccess)

	var listCreation func(int) *Node
	listCreation = randomValueList
	if *addressOrderedList {
		listCreation = memoryOrderedList
	}
	if *alreadySorted {
		listCreation = presortedList
	}

	for n := *countBegin; n < *countUntil; n += *countIncrement {
		var total time.Duration
		var looping time.Duration
		var head *Node
		for i := 0; i < 10; i++ {
			head = listCreation(n)

			var nl *Node
			before := time.Now()
			nl = buMergesort(head, mergeFn)
			elapsed := time.Since(before)
			total += elapsed

			if sz := listSize(nl); sz != n {
				log.Printf("list of size %d had %d elements after sort\n", n, sz)
				os.Exit(2)
			}

			elapsed = time.Since(before)
			looping += elapsed
		}
		total /= 10.0
		fmt.Printf("%d\t%.04f\t%.04f\n", n, total.Seconds(), looping.Seconds())
	}

	fmt.Printf("# ending at %s on %s\n", time.Now().Format(time.RFC3339), hostname)
}

func listSize(head *Node) int {
	if head == nil {
		return 0
	}
	if head.Next == nil {
		return 1
	}
	var sz int
	for ; head != nil; head = head.Next {
		sz++
	}
	return sz
}

func randomValueList(n int) *Node {

	var head *Node

	for i := 0; i < n; i++ {
		head = &Node{
			Data: uint(rand.Int()),
			Next: head,
		}
	}

	return head
}

func presortedList(n int) *Node {

	var head *Node

	for i := 0; i < n; i++ {
		head = &Node{
			Data: uint(i),
			Next: head,
		}
	}

	return head
}

func memoryOrderedList(n int) *Node {

	head := &Node{}
	head.Data = uint(uintptr(unsafe.Pointer(head)))
	tail := head

	// Append new *Node to end of list - this will create
	// a list that has blocks of nodes in descending address order
	for i := 1; i < n; i++ {
		nn := &Node{}
		nn.Data = uint(uintptr(unsafe.Pointer(nn)))
		tail.Next = nn
		tail = tail.Next
	}

	// sort all nodes by address, so that even blocks of nodes are
	// ordered by ascending address.
	head = recursiveMergeSort(head)
	return rerandomizeList(head)
}

func rerandomizeList(head *Node) *Node {
	for node := head; node != nil; node = node.Next {
		node.Data = uint(rand.Int())
	}
	return head
}

func mergesort(head *Node) *Node {

	var hd, tl *Node
	appnd := func(n *Node) {
		if hd == nil {
			hd = n
			tl = n
			return
		}
		tl.Next = n
		tl = n
	}

	p := head
	mergecount := 2 // just to pass the first for-test

	// The final pass over the unsorted linked list merges
	// two lists each of about half the number of nodes.
	// mergecount will have value 1 in that case. Don't
	// need to loop again.
	for k := 1; mergecount > 1; k *= 2 {

		mergecount = 0

		for p != nil {

			psize := 0
			q := p
			for i := 0; q != nil && i < k; i++ {
				psize++
				q = q.Next
			}

			qsize := psize

			for psize > 0 && qsize > 0 && q != nil {
				if p.Data < q.Data {
					appnd(p)
					p = p.Next
					psize--
					continue
				}
				appnd(q)
				q = q.Next
				qsize--
			}

			for ; psize > 0 && p != nil; psize-- {
				appnd(p)
				p = p.Next
			}

			for ; qsize > 0 && q != nil; qsize-- {
				appnd(q)
				q = q.Next
			}

			p = q

			mergecount++
		}

		p = hd
		head = hd

		hd = nil
		tl.Next = nil
		tl = nil
	}

	return head
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
		n := &right
		if left.Data < right.Data {
			n = &left
		}
		t.Next = *n
		*n = (*n).Next
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

// buMergesort - transliteration of Wikipedia's "Bottom up implementation with lists",
// https://en.wikipedia.org/wiki/Merge_sort#Bottom-up_implementation_using_lists
func buMergesort(head *Node, nodeFn MergeFn) *Node {
	if head == nil {
		return nil
	}
	// Can pass 1-length lists to the rest of the function,
	// because array[0] is 2^0 or 1 in size

	var array [32]*Node
	var result, next *Node
	var i int

	result = head

	for result != nil {
		next = result.Next
		result.Next = nil

		for i = 0; (i < 32) && (array[i] != nil); i++ {
			result = nodeFn(array[i], result)
			array[i] = nil
		}
		if i == 32 {
			i--
		}
		array[i] = result
		result = next
	}

	result = nil
	for i = 0; i < 32; i++ {
		result = nodeFn(array[i], result)
	}

	return result
}

// weaveSublists alternates elements of p and q,
// it walks both lists and updates almost all .Next pointers
func weaveSublists(p *Node, q *Node) *Node {
	if p == nil {
		return q
	}
	if q == nil {
		return p
	}

	x := &p
	y := &q
	if rand.Intn(2) == 0 {
		x = &q
		y = &p
	}

	h, t := *x, *x
	*x = (*x).Next

	for *x != nil && *y != nil {
		t.Next = *y
		t = t.Next
		*y = (*y).Next

		t.Next = *x
		t = t.Next
		*x = (*x).Next
	}

	for *x != nil {
		t.Next = *x
		t = t.Next
		*x = (*x).Next
	}

	for *y != nil {
		t.Next = *y
		t = t.Next
		*y = (*y).Next
	}

	t.Next = nil

	return h
}

// incrementData walks both lists and increments the .Data element.
// No .Next pointer updates or changes
func incrementData(p *Node, q *Node) *Node {
	if p == nil {
		return q
	}
	if q == nil {
		return p
	}

	var tmp *Node
	for tmp = p; tmp.Next != nil; tmp = tmp.Next {
		tmp.Data++
	}
	// tmp not nil, but tmp.Next is.
	tmp.Data++

	// I lied about no .Next updates. There's one, it appends q to p
	tmp.Next = q
	for tmp = q; tmp != nil; tmp = tmp.Next {
		tmp.Data++
	}

	return p
}

func Print(list *Node) {
	for node := list; node != nil; node = node.Next {
		fmt.Printf("%d -> ", node.Data)
	}
	fmt.Println()
}
