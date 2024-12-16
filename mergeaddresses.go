package main

// Show the addresses of heads of merged lists,
// same merged list in recursive and wikipedia bottom up

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
	countBegin := flag.Int("b", 64, "beginning list size")
	flag.Parse()

	rand.Seed(time.Now().UnixNano() | int64(os.Getpid()))
	hostname, _ := os.Hostname() // not going to fail
	fmt.Printf("# %s on %s\n", time.Now().Format(time.RFC3339), hostname)
	fmt.Printf("# List of %d nodes\n", *countBegin)
	fmt.Printf("# nodes %d bytes in size\n", unsafe.Sizeof(Node{}))

	// 1. Create a list with randomly-chosen integer data values
	// 2. Print the list, which shows data values
	// 3. Do recursive merge sort on the list, printing out
	//    addresses and sizes of lists to merge
	// 4. Check size and sortedness of list
	// 5. Reset list to original node ordering
	// 6. Print the list again, to allow check of list reset
	// 7. Sort list via wikipedia bottom up algorithm,
	//    printing out addresses and sizes of lists to merge
	// 8. Check size and sortedness of list

	head := randomValueList(*countBegin, *countBegin)

	var nl *Node
	var sz int
	fmt.Printf("# recursive sort\n")
	fmt.Print("# ")
	Print(head)
	fmt.Println()
	nl = recursiveMergeSort(head)
	fmt.Printf("# %d comparisons during recursive merge sort of %d length list\n", recursiveComparisonCount, *countBegin)
	checkSorted(nl, *countBegin, "first")
	sz = resetList(head)
	if sz != *countBegin {
		log.Fatalf("list should be %d length, is %d after list reset\n", *countBegin, sz)
	}
	fmt.Printf("# bottom up sort\n")
	fmt.Print("# ")
	Print(head)
	fmt.Println()
	nl = buMergesort(head)
	fmt.Printf("# %d comparisons during bottom up merge sort of %d length list\n", buComparisonCount, *countBegin)
	checkSorted(nl, *countBegin, "second")

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
func resetList(head *Node) int {
	length := 0
	for node := head; node != nil; node = node.Reset {
		node.Next = node.Reset
		length++
	}
	return length
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

// randomValueList creates a linked list with randomly-chosen
// integer data values, up to max in size.
// Two chains through the list, one via .Next pointers,
// the other through .Reset pointers, which are used to reconstruct
// the original order of the list
func randomValueList(n int, max int) *Node {

	var head *Node

	for i := 0; i < n; i++ {
		head = &Node{
			Data:  uint(rand.Intn(max)),
			Next:  head,
			Reset: head, // original order of nodes restorable
		}
	}

	return head
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

	leftLength, _ := isSorted(left)
	rightLength, _ := isSorted(right)

	fmt.Printf("merging <%d,%d> (%p, %p)\n", leftLength, rightLength, left, right)

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

// Print runs a linked list and prints its values on stdout
func Print(list *Node) {
	for node := list; node != nil; node = node.Next {
		fmt.Printf("%d -> ", node.Data)
	}
	fmt.Println()
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
			resultLength, _ := isSorted(result)
			iLength, _ := isSorted(array[i])
			fmt.Printf("merging <%d,%d> (%p, %p)\n", resultLength, iLength, array[i], result)
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
		resultLength, _ := isSorted(result)
		iLength, _ := isSorted(array[i])
		phrase := "nil"
		if array[i] != nil {
			phrase = "non-nil"
		}
		fmt.Printf("%d %s, merging <%d,%d> (%p, %p)\n", i, phrase, resultLength, iLength, array[i], result)
		result = merge(array[i], result)
	}

	return result
}
