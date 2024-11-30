package main

/* Recursive mergesort a few odd ways */

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
	useRecursiveSort2 := flag.Bool("z", false, "use purely recursive mergesort with user stack")
	useRecursiveSort3 := flag.Bool("a", false, "use purely recursive alternating mergesort")
	useRecursiveSort4 := flag.Bool("d", false, "use purely recursive mergesort, rhs first")
	useRecursiveSort5 := flag.Bool("e", false, "use counted purely recursive mergesort")
	useRecursiveSort6 := flag.Bool("f", false, "recursive mergesort with merge function")
	useRecursiveSort7 := flag.Bool("o", false, "recursive mergesort with user stack 2")
	useRecursiveSort8 := flag.Bool("p", false, "recursive mergesort with user stack 3")
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
	hostname, _ := os.Hostname() // not going to fail
	fmt.Printf("# %s on %s\n", time.Now().Format(time.RFC3339), hostname)
	fmt.Printf("# Start at %d nodes, end before %d nodes, increment %d\n",
		*countBegin, *countUntil, *countIncrement)
	sortType := "unknown"
	if *useRecursiveSort {
		sortType = "recursive"
	} else if *useRecursiveSort2 {
		sortType = "recursive with user-level stack"
	} else if *useRecursiveSort3 {
		sortType = "recursive with alternating splits"
	} else if *useRecursiveSort4 {
		sortType = "recursive rhs first"
	} else if *useRecursiveSort5 {
		sortType = "counted recursive"
	} else if *useRecursiveSort6 {
		sortType = "recursive, with merge function"
	} else if *useRecursiveSort7 {
		sortType = "recursive, with user stack 2"
	} else if *useRecursiveSort8 {
		sortType = "recursive, with user stack 3"
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
		listCreationPhrase = "reverse sorted"
	}
	fmt.Printf("# %s data values\n", listCreationPhrase)

	for n := *countBegin; n < *countUntil; n += *countIncrement {
		var total time.Duration
		var looping time.Duration
		var head *Node
		if *reuseList {
			head = listCreation(n, *useCryptoRand)
		}
		min := time.Duration((365 * 24 * 3600) * time.Second)
		max := time.Duration(0)
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
			case *useRecursiveSort2:
				nl = ownstackMergeSort(head)
			case *useRecursiveSort3:
				nl = recursiveMergeSortLeft(head)
			case *useRecursiveSort4:
				nl = recursiveMergeSortRHS(head)
			case *useRecursiveSort5:
				nl = countedRecursiveMergeSort(head, n)
			case *useRecursiveSort6:
				nl = recursiveMergeSortMerge(head)
			case *useRecursiveSort7:
				nl = ownstackMergeSort2(head)
			case *useRecursiveSort8:
				nl = ownstackMergeSort3(head)
			}
			elapsed := time.Since(before)
			total += elapsed
			if elapsed > max {
				max = elapsed
			}
			if elapsed < min {
				min = elapsed
			}

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
		fmt.Printf("%d\t%.04f\t%.04f\t%.04f\t%.04f\n", n, total.Seconds(), looping.Seconds(), min.Seconds(), max.Seconds())
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

func recursiveMergeSortMerge(head *Node) *Node {
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
	*turtle = nil // nil-terminates left side sublist

	left := recursiveMergeSortMerge(head)
	right = recursiveMergeSortMerge(right)

	return merge(left, right)
}

// countedRecursiveMergeSort same as func recursiveMergeSort,
// except it only touches half the list nodes to find
// the middle of the list.
func countedRecursiveMergeSort(head *Node, size int) *Node {
	if head.Next == nil {
		// single node list is sorted by definition
		return head
	}

	// because of recursion bottoming out at a 1-long-list,
	// head points to a list of at least 2 elements.

	p := &head
	var i, leftSize int

	for i = 0; i < size; i += 2 {
		p = &((*p).Next)
		leftSize++
	}

	right := *p
	*p = nil

	left := countedRecursiveMergeSort(head, leftSize)
	right = countedRecursiveMergeSort(right, size-leftSize)

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

func recursiveMergeSortRHS(head *Node) *Node {
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

	left := recursiveMergeSortRHS(right)
	right = recursiveMergeSortRHS(head)

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

type stackFrame struct {
	list   *Node
	merged *Node
	next   *stackFrame
}

func split(head *Node) (*Node, *Node) {
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
	return head, right
}

func ownstackMergeSort(head *Node) *Node {

	stack := &stackFrame{
		list: head,
	}

	var sorted *Node

	for {
		var elem *stackFrame

		elem, stack = stack, stack.next

		if elem.list == nil && stack == nil {
			sorted = elem.merged
			break
		}

		if elem.list != nil && elem.list.Next == nil {
			// "recursion" has bottomed out at 1-node list
			elem.merged, elem.list = elem.list, nil
			elem.next, stack = stack, elem
			continue
		}

		if elem.merged != nil {
			// a merged sublist has "returned"

			tmp := stack
			stack = stack.next

			if tmp.merged == nil {
				elem.next, stack = stack, elem
				tmp.next, stack = stack, tmp
				continue
			}

			// both tmp and elem contain merged sublists

			elem.merged = merge(elem.merged, tmp.merged)
			elem.next, stack = stack, elem
			// discarding tmp
			continue
		}

		// still "recursing"
		left, right := split(elem.list)
		stack = &stackFrame{
			list: left,
			next: stack,
		}
		stack = &stackFrame{
			list: right,
			next: stack,
		}
	}

	return sorted
}

type stackFrame2 struct {
	formalArgument *Node
	left           *Node
	right          *Node
	leftSorted     *Node
	next           *stackFrame2
}

// ownstackMergeSort2 - a "recursive" mergesort that's all user level.
// There's no implicit function call stack, it's explicit. The "stack frame"
// is struct stackFrame2. Further, both "stack" and "stack frames" are
// allocated on the call stack of func ownstackMergeSort2. I think there's
// no heap allocations in this function.
func ownstackMergeSort2(head *Node) *Node {

	if head == nil {
		return nil
	}

	var stack [32]stackFrame2
	var ply int
	var returnValue *Node

	stack[ply].formalArgument = head

	for {
		if stack[ply].formalArgument.Next == nil && stack[ply].left == nil {
			// "recursion" has bottomed out
			// fmt.Printf("1-length list, recursion bottomed-out\n")
			returnValue = stack[ply].formalArgument
			// .left, .right, .leftSorted should all contain nil
			ply--
			continue // "return" from bottomed-out recursion
		}

		if stack[ply].leftSorted == nil {
			if stack[ply].left == nil {
				// haven't recursed on either .left or .right
				stack[ply].left, stack[ply].right = split(stack[ply].formalArgument)
				// set up stack frame for mergesort(left)
				tmp := stack[ply].left
				ply++ // stack[ply] is a new frame
				stack[ply].formalArgument = tmp
				continue // "call" mergesort(left)
			}
			// returned from mergesort(left), returnValue should not contain nil
			stack[ply].leftSorted = returnValue
			returnValue = nil
			// set up stack frame for mergesort(right)
			tmp := stack[ply].right
			ply++
			stack[ply].formalArgument = tmp
			continue // "call" mergesort(right)
		}

		// stack[ply].leftSorted != nil, "return" from mergesort(right)

		returnValue = merge(stack[ply].leftSorted, returnValue)
		stack[ply].leftSorted = nil
		stack[ply].left = nil
		stack[ply].right = nil
		ply--
		if ply < 0 {
			break
		}
		// returnValue is non-nil, the merge of .left and .right
	}

	return returnValue
}

func ownstackMergeSort3(head *Node) *Node {

	if head == nil {
		return nil
	}

	var frames *stackFrame2

	for i := 0; i < 32; i++ {
		frame := new(stackFrame2)
		frame.next = frames
		frames = frame
	}

	return realOwnstackMergeSort3(head, frames)
}

func realOwnstackMergeSort3(head *Node, frames *stackFrame2) *Node {

	var stack *stackFrame2
	var returnValue *Node

	frame := frames
	frames = frames.next

	frame.formalArgument = head

	frame.next = stack
	stack = frame

	for {

		frame = stack
		stack = stack.next

		if frame.formalArgument.Next == nil && frame.left == nil {
			// "recursion" has bottomed out
			// fmt.Printf("1-length list, recursion bottomed-out\n")
			returnValue = frame.formalArgument
			// .left, .right, .leftSorted should all contain nil
			frame.next = frames
			frames = frame
			continue // "return" from bottomed-out recursion
		}

		if frame.leftSorted == nil {
			if frame.left == nil {
				// haven't recursed on either .left or .right
				frame.left, frame.right = split(frame.formalArgument)
				// set up stack frame for mergesort(left)
				newframe := frames
				frames = frames.next
				newframe.formalArgument = frame.left
				frame.next = stack
				stack = frame
				newframe.next = stack
				stack = newframe
				continue // "call" mergesort(left)
			}
			// returned from mergesort(left), returnValue should not contain nil
			frame.leftSorted = returnValue
			returnValue = nil
			// set up stack frame for mergesort(right)
			newframe := frames
			frames = frames.next
			newframe.formalArgument = frame.right
			frame.next = stack
			stack = frame
			newframe.next = stack
			stack = newframe
			continue // "call" mergesort(right)
		}

		// frame.leftSorted != nil, "return" from mergesort(right)

		returnValue = merge(frame.leftSorted, returnValue)
		frame.leftSorted = nil
		frame.left = nil
		frame.right = nil
		frame.next = frames
		frames = frame
		if stack == nil {
			break
		}
		// returnValue is non-nil, the merge of .left and .right
	}

	return returnValue
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
func recursiveMergeSortLeft(head *Node) *Node {
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

	left := recursiveMergeSortRight(right)
	right = recursiveMergeSortRight(head)

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
func recursiveMergeSortRight(head *Node) *Node {
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

	left := recursiveMergeSortLeft(head)
	right = recursiveMergeSortLeft(right)

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

func listSize(node *Node) int {
	count := 0
	for ; node != nil; node = node.Next {
		count++
	}
	return count
}
