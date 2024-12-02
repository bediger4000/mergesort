package main

/*
 * Count sorting comparisons
 */

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
	Data  uint
	Reset *Node // another chain through the linked list
	Next  *Node
}

func main() {
	alreadySorted := flag.Bool("s", false, "already sorted low-to-high list")
	reverseSorted := flag.Bool("S", false, "reverse sorted high-to-low list")

	countIncrement := flag.Int("i", 200000, "increment of list size")
	countBegin := flag.Int("b", 1000, "beginning list size")
	countUntil := flag.Int("u", 18000000, "sort lists up to this size")

	flag.Parse()

	rand.Seed(time.Now().UnixNano() | int64(os.Getpid()))
	hostname, _ := os.Hostname() // not going to fail

	fmt.Printf("# %s on %s\n", time.Now().Format(time.RFC3339), hostname)
	fmt.Printf("# Start at %d nodes, end before %d nodes, increment %d\n",
		*countBegin, *countUntil, *countIncrement)

	fmt.Print("# idiomatic list in-memory ordering\n")
	fmt.Print("# math/rand random numbers as list node values\n")
	fmt.Printf("# nodes %d bytes in size\n", unsafe.Sizeof(Node{}))

	var listCreation func(int) *Node
	listCreation = randomValueList
	listCreationPhrase := "randomly chosen data"
	if *alreadySorted {
		listCreation = presortedList
		listCreationPhrase = "presorted"
	}
	if *reverseSorted {
		listCreation = reverseSortedList
		listCreationPhrase = "reverse sorted"
	}
	fmt.Printf("# %s data values\n", listCreationPhrase)

	for n := *countBegin; n < *countUntil; n += *countIncrement {

		head := listCreation(n)

		var nl *Node

		// Zero out global comparison counters
		recursiveComparisonCount = 0
		buComparisonCount = 0
		iterativeComparisonCount = 0

		// recursive mergesort, check list
		nl = recursiveMergeSort(head)
		checkSorted(nl, n, "recursive")
		resetList(head, n, "recursive")

		// bottom up mergesort, check list
		nl = buMergesort(head)
		checkSorted(nl, n, "bottom up")
		resetList(head, n, "bottom up")

		// iterative mergesort, check list
		nl = mergesort(head)
		checkSorted(nl, n, "iterative")

		fmt.Printf("%d\t%d\t%d\t%d\n", n, recursiveComparisonCount, buComparisonCount, iterativeComparisonCount)
	}

	fmt.Printf("# ending at %s on %s\n", time.Now().Format(time.RFC3339), hostname)
}

func checkSorted(head *Node, nominalSize int, phrase string) {
	if sz, sorted := isSorted(head); !sorted {
		log.Printf("list of size %d not sorted at element %d, %s\n", nominalSize, sz, phrase)
		os.Exit(1)
	} else if sz != nominalSize {
		log.Printf("list of size %d had %d elements after %s sort\n", nominalSize, sz, phrase)
		os.Exit(2)
	}
}

// resetList walks the list that starts with head
// via .Reset pointers. It sets .Next to .Reset
// in every list node, restoring the original node
// order.
func resetList(head *Node, nominalLength int, phrase string) {
	length := 0
	for node := head; node != nil; node = node.Reset {
		node.Next = node.Reset
		length++
	}

	if length != nominalLength {
		log.Printf("reset after %s sort, reset list %d nodes, should have been %d\n",
			phrase, length, nominalLength)
		os.Exit(3)
	}
}

func isSorted(head *Node) (int, bool) {
	if head == nil {
		return 0, true
	}
	if head.Next == nil {
		return 1, true
	}
	var sz int
	for ; head.Next != nil; head = head.Next {
		sz++
		if head.Data > head.Next.Data {
			return sz, false
		}
	}
	sz++ // for-loop checks head.Next, count final element on list
	return sz, true
}

func randomValueList(n int) *Node {

	var head *Node

	for i := 0; i < n; i++ {
		head = &Node{
			Data:  uint(rand.Int()),
			Next:  head,
			Reset: head,
		}
	}

	return head
}

func presortedList(n int) *Node {

	var head *Node

	for i := n - 1; i >= 0; i-- {
		head = &Node{
			Data:  uint(i),
			Next:  head,
			Reset: head,
		}
	}

	return head
}

func reverseSortedList(n int) *Node {

	var head *Node

	for i := 0; i < n; i++ {
		head = &Node{
			Data:  uint(i),
			Next:  head,
			Reset: head,
		}
	}

	return head
}

// Print runs a linked list and prints its values on stdout
func Print(list *Node) {
	for node := list; node != nil; node = node.Next {
		fmt.Printf("%d -> ", node.Data)
	}
	fmt.Println()
}

func listSize(node *Node) int {
	count := 0
	for ; node != nil; node = node.Next {
		count++
	}
	return count
}

var recursiveComparisonCount int

func recursiveMergeSort(head *Node) *Node {
	if head.Next == nil {
		// single node list is sorted by definition
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
	recursiveComparisonCount++
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
		recursiveComparisonCount++
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

var buComparisonCount int

// buMergesort - transliteration of Wikipedia's "Bottom up implementation with lists",
// https://en.wikipedia.org/wiki/Merge_sort#Bottom-up_implementation_using_lists
func buMergesort(head *Node) *Node {
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
			result = merge(array[i], result)
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
		result = merge(array[i], result)
	}

	return result
}

// merge for bottom up - it counts comparisons for that algorithm
func merge(p *Node, q *Node) *Node {
	if p == nil {
		return q
	}
	if q == nil {
		return p
	}

	x := &q
	buComparisonCount++
	if p.Data < q.Data {
		x = &p
	}

	h, t := *x, *x
	*x = (*x).Next

	for p != nil && q != nil {
		n := &q
		buComparisonCount++
		if p.Data < q.Data {
			n = &p
		}
		t.Next = *n
		*n = (*n).Next
		t = t.Next
	}

	t.Next = p
	if q != nil {
		t.Next = q
	}

	return h
}

var iterativeComparisonCount int

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
				iterativeComparisonCount++
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
