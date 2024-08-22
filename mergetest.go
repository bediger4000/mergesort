package main

import (
	crand "crypto/rand"
	"flag"
	"fmt"
	"log"
	"math"
	"math/big"
	"math/rand"
	"os"
	"runtime"
	"time"
	"unsafe"
)

// Node is an element of a linked list
type Node struct {
	Data uint
	Next *Node
}

func main() {
	useCryptoRand := flag.Bool("c", false, "use cryptographic PRNG")
	useRecursiveSort := flag.Bool("r", false, "use purely recursive mergesort")
	useBottomUp := flag.Bool("B", false, "bottom-up mergesort with lists")
	reuseList := flag.Bool("R", false, "re-randomize and re-use list")
	alreadySorted := flag.Bool("s", false, "already sorted low-to-high list")
	reverseSorted := flag.Bool("S", false, "reverse sorted high-to-low list")
	addressOrderedList := flag.Bool("m", false, "create address-ordered list for each sort")
	garbageCollectAfter := flag.Bool("G", false, "collect garbage after each sort")
	countIncrement := flag.Int("i", 200000, "increment of list size")
	countBegin := flag.Int("b", 1000, "beginning list size")
	countUntil := flag.Int("u", 18000000, "sort lists up to this size")
	flag.Parse()

	rand.Seed(time.Now().UnixNano() | int64(os.Getpid()))
	if *useRecursiveSort && *useBottomUp {
		log.Fatalf("only one of -r and -B allowed\n")
	}
	hostname, _ := os.Hostname() // not going to fail
	fmt.Printf("# %s on %s\n", time.Now().Format(time.RFC3339), hostname)
	fmt.Printf("# Start at %d nodes, end before %d nodes, increment %d\n",
		*countBegin, *countUntil, *countIncrement)
	sortType := "iterative"
	if *useRecursiveSort {
		sortType = "recursive"
	} else if *useBottomUp {
		sortType = "bottom-up iterative"
	}
	fmt.Printf("# %s sort\n", sortType)
	listType := "idomatic"
	if *addressOrderedList {
		listType = "memory address"
	}
	fmt.Printf("# %s list in-memory ordering\n", listType)
	if *reuseList {
		fmt.Println("# re-random-value and re-use list")
	}
	if *garbageCollectAfter {
		fmt.Println("# garbage collect after each sort iteration")
	}
	randomType := "math/rand"
	if *useCryptoRand {
		randomType = "cryptographic"
	}
	fmt.Printf("# %s random numbers as list node values\n", randomType)
	fmt.Printf("# nodes %d bytes in size\n", unsafe.Sizeof(Node{}))

	var listCreation func(int, bool) *Node
	listCreation = randomValueList
	listCreationPhrase := "randomly chosen data"
	if *addressOrderedList {
		listCreation = memoryOrderedList
		listCreationPhrase = "unordered"
		fmt.Printf("# node addresses ascending in memory\n")
	}
	if *alreadySorted {
		listCreation = presortedList
		listCreationPhrase = "presorted"
	}
	if *reverseSorted {
		listCreation = reverseSortedList
		listCreationPhrase = "revese sorted"
	}
	fmt.Printf("# %s data values\n", listCreationPhrase)

	for n := *countBegin; n < *countUntil; n += *countIncrement {
		var total time.Duration
		var looping time.Duration
		var head *Node
		if *reuseList {
			head = listCreation(n, *useCryptoRand)
		}
		for i := 0; i < 10; i++ {
			beforeIteration := time.Now()
			if !*reuseList {
				// fresh, new list every iteration
				head = listCreation(n, *useCryptoRand)
			}

			var nl *Node
			before := time.Now()
			switch {
			case *useRecursiveSort:
				nl = recursiveMergeSort(head)
			case *useBottomUp:
				nl = buMergesort(head)
			default:
				nl = mergesort(head)
			}
			elapsed := time.Since(before)
			total += elapsed

			if sz, sorted := isSorted(nl); !sorted {
				log.Printf("list of size %d not sorted at element %d\n", n, sz)
				os.Exit(1)
			} else if sz != n {
				log.Printf("list of size %d had %d elements after sort\n", n, sz)
				os.Exit(2)
			}

			if *reuseList {
				head = rerandomizeList(nl, *useCryptoRand)
			}

			if *garbageCollectAfter {
				head = nil
				nl = nil
				runtime.GC()
			}
			elapsed = time.Since(beforeIteration)
			looping += elapsed
		}
		total /= 10.0
		fmt.Printf("%d\t%.04f\t%.04f\n", n, total.Seconds(), looping.Seconds())
	}

	fmt.Printf("# ending at %s on %s\n", time.Now().Format(time.RFC3339), hostname)
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

var maxInt = big.NewInt(math.MaxInt32)

func randomValueList(n int, useCheapRand bool) *Node {

	var head *Node

	for i := 0; i < n; i++ {
		head = &Node{
			Data: randomValue(useCheapRand),
			Next: head,
		}
	}

	return head
}

func presortedList(n int, _ bool) *Node {

	var head *Node

	for i := n - 1; i >= 0; i-- {
		head = &Node{
			Data: uint(i),
			Next: head,
		}
	}

	return head
}

func reverseSortedList(n int, _ bool) *Node {

	var head *Node

	for i := 0; i < n; i++ {
		head = &Node{
			Data: uint(i),
			Next: head,
		}
	}

	return head
}

func memoryOrderedList(n int, useCheapRand bool) *Node {

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
	return rerandomizeList(head, useCheapRand)
}

func rerandomizeList(head *Node, useCheapRand bool) *Node {
	for node := head; node != nil; node = node.Next {
		node.Data = randomValue(useCheapRand)
	}
	return head
}

func randomValue(useCheapRand bool) uint {
	var ri int
	if useCheapRand {
		ri = rand.Int()
	} else {
		mp, err := crand.Int(crand.Reader, maxInt)
		if err != nil {
			log.Fatal(err)
		}
		ri = int(mp.Int64())
	}
	return uint(ri)
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

func merge(p *Node, q *Node) *Node {
	if p == nil {
		return q
	}
	if q == nil {
		return p
	}

	x := &q
	if p.Data < q.Data {
		x = &p
	}

	h, t := *x, *x
	*x = (*x).Next

	for p != nil && q != nil {
		n := &q
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

// Print runs a linked list and prints its values on stdout
func Print(list *Node) {
	for node := list; node != nil; node = node.Next {
		fmt.Printf("%d -> ", node.Data)
	}
	fmt.Println()
}
