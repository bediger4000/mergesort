# Merge sort of linked lists

A program to benchmark mergesort of various length linked lists.

## Building and Running

To build:

```
$ go build mergetest.go
```

`mergetest` has a variety of command line options:

```
  -B    bottom-up mergesort with lists
  -G    collect garbage after each sort
  -R    re-randomize and re-use list
  -S    reverse sorted high-to-low list
  -b int
        beginning list size (default 1000)
  -c    use cryptographic PRNG
  -i int
        increment of list size (default 200000)
  -m    create address-ordered list for each sort
  -r    use purely recursive mergesort
  -s    already sorted low-to-high list
  -u int
        sort lists up to this size (default 18000000)
  -z    use recursive mergesort with user stack
```

You can consider
`mergetest` to have four sets of options.

### Setting linked list sizes and increments

The `mergetest` program runs multiple linked list sizes per run.
You can control what sizes of linked lists are sorted,
and how big the increment in linked lists between timing sets.

- `-b` starting linked list length, default 1,000 nodes
- `-u` sort linked lists up to this list length, default 18,000,000 nodes
- `-i` increment linked list size by this amount between timed sets of sortings, default 200,000 nodes

### Setting numerical value of linked list nodes

- By default, use pseudo-random number generator to create unsorted linked lists
- `-s` created sorted linked lists, node data values zero to max
- `-S` created reverse sorted linked lists, node data values max to zero

The mergesort variants all sort node data values low-to-high.

### Choosing method of setting numerical value of unsorted linked list nodes

- By default, using Go's `math/rand` non-cryptographic pseudo-random number generator
to create an unsorted linked list
- `-c` Use Go's `crypto/rand` cryptographically-strong  pseudo-random number generator

### Select mergesort variant

- By default, my own iterative mergesort using O(1) extra space
- `-B` Wikipedia's "bottom up" iterative mergesort
- `-r` purely recursive mergesort
- `-z` recursive mergesort with user-level stack

### Arrange the initial linked list in memory

- By default, allocate linked list nodes "idiomatically"
- `-m` order linked list nodes from low memory to high

### Miscellaneous Options

- `-G` collect garbage after each sort
- `-R` don't re-create a linked list, re-randomize node values and re-use

You can use `-G` with any assortment of other options.
Using `-R` will cause `mergetest` to create linked lists using
whatever option (default, `-m`) the first of 10 sorts.
For the other 9 sortings, the code runs the sorted linked list
by `.Next` pointer and assigned random numerical values to the
nodes' `.Data` fields.
