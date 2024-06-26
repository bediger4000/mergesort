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
	"time"
)

// Node is an element of a linked list
type Node struct {
	Data int
	Next *Node
}

func main() {
	useCryptoRand := flag.Bool("c", false, "use cryptographic PRNG")
	flag.Parse()

	rand.Seed(time.Now().UnixNano() | int64(os.Getpid()))

	for n := 10000; n < 8000000; n += 200000 {
		var total time.Duration
		for i := 0; i < 10; i++ {
			head := randomValueList(n, *useCryptoRand)

			before := time.Now()
			nl := mergesort(head)
			elapsed := time.Since(before)

			if !isSorted(nl) {
				os.Exit(1)
			}
			total += elapsed

		}
		total /= 10.0
		fmt.Printf("%d\t%.04f\n", n, total.Seconds())
	}
}

func isSorted(head *Node) bool {
	if head == nil || head.Next == nil {
		return true
	}
	for ; head.Next != nil; head = head.Next {
		if head.Data > head.Next.Data {
			return false
		}
	}
	return true
}

var maxInt = big.NewInt(math.MaxInt32)

func randomValueList(n int, useCheapRand bool) *Node {

	var head *Node

	for i := 0; i < n; i++ {
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
		head = &Node{
			Data: ri,
			Next: head,
		}
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
