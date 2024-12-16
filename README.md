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

## Output

`mergetest` produces output that is easy to use in `gnuplot`.
Typical output looks like this:

```
# 2024-09-29T21:48:14-06:00 on modest
# Start at 1000 nodes, end before 18000000 nodes, increment 40000
# recursive sort
# idomatic list in-memory ordering
# math/rand random numbers as list node values
# nodes 16 bytes in size
# randomly chosen data data values
1000    0.0002  0.0324  0.0001  0.0002
41000   0.0100  1.2179  0.0098  0.0103
81000   0.0237  2.4883  0.0216  0.0294
121000  0.0424  3.9718  0.0369  0.0562
...
17881000    23.4298 780.2499    23.1556 23.6560
17921000    23.8185 786.3332    22.9558 24.1528
17961000    23.6769 788.9641    23.0547 24.2352
# ending at 2024-10-01T21:08:12-06:00 on modest
```

Comment/provenance lines begin with '#'.
Data lines are all others, each consisting of 5, tab-separated, numerical values.

1. Linked list length, number of nodes
2. Arithmetic mean of 10 sorts of linked lists of that length, seconds
3. Total elapsed time, included list set up, for those 10 sorts, seconds
4. Minimum elapsed time of the 10 sorts, seconds
5. Maximum elapsed time of the 10 sorts, seconds

## Check the order in which two algorithms access memory

`mergeaddresses.go` sorts the same list with recursive and
wikipedia's bottom up algorithm,
displaying merging lists' lengths and addresses of head nodes.

```
$ go build mergeaddresses.go
$ ./mergeaddresses -b 8
# 2024-11-21T21:42:36-07:00 on hazard
# List of 8 nodes
# nodes 24 bytes in size
# recursive sort
# 5 -> 0 -> 0 -> 7 -> 1 -> 0 -> 6 -> 1 -> 

merging <1,1> (0xc0001160a8, 0xc000116090)
merging <1,1> (0xc000116078, 0xc000116060)
merging <2,2> (0xc000116090, 0xc000116078)
merging <1,1> (0xc000116048, 0xc000116030)
merging <1,1> (0xc000116018, 0xc000116000)
merging <2,2> (0xc000116030, 0xc000116000)
merging <4,4> (0xc000116078, 0xc000116030)
# bottom up sort
# 5 -> 0 -> 0 -> 7 -> 1 -> 0 -> 6 -> 1 -> 

merging <1,1> (0xc0001160a8, 0xc000116090)
merging <1,1> (0xc000116078, 0xc000116060)
merging <2,2> (0xc000116090, 0xc000116078)
merging <1,1> (0xc000116048, 0xc000116030)
merging <1,1> (0xc000116018, 0xc000116000)
merging <2,2> (0xc000116030, 0xc000116000)
merging <4,4> (0xc000116078, 0xc000116030)
# ending at 2024-11-21T21:42:36-07:00 on hazard
```

These two algorithms read and write list nodes
in the same order when sorting.

## Recursive mergesort algorithm variations

```
$ go build recursivetest.go

Usage of ./recursivetest:
  -G    collect garbage after each sort
  -R    re-randomize and re-use list
  -S    reverse sorted high-to-low list
  -a    use purely recursive alternating mergesort
  -b int
        beginning list size (default 1000)
  -c    use cryptographic PRNG
  -d    use purely recursive mergesort, rhs first
  -e    use counted purely recursive mergesort
  -f    recursive mergesort with merge function
  -i int
        increment of list size (default 200000)
  -m    create address-ordered list for each sort
  -o    recursive mergesort with user stack 2
  -p    recursive mergesort with user stack 3
  -r    use purely recursive mergesort
  -s    already sorted low-to-high list
  -u int
        sort lists up to this size (default 18000000)
  -z    use purely recursive mergesort with user stack
```

The major variations are:

* -a  recursive list sort, recursion alternating left and right lists
* -d  recursive list sort from end of list instead of head
* -e  counted list splits instead of walking the formal argument list
* -f  list merges in a function, rather than in-line
* -o  userland simulated call stack allocated on the function call stack
* -p  userland simulated call stack allocated on the process heap

## Mergesort comparison counting

```
$ go build cmpcounter2.go 

Usage of ./cmpcounter2:
  -I int
        number of sorts conducted at any given list length (default 10)
  -S    reverse sorted high-to-low list
  -b int
        beginning list size (default 1000)
  -i int
        increment of list size (default 200000)
  -s    already sorted low-to-high list
  -u int
        sort lists up to this size (default 18000000)
```

Counts the number of `if left.data < right.data` comparisons done to sort a list.
Output is a little different:

```
# 2024-12-15T18:15:32-07:00 on hazard
# Start at 1000 nodes, end before 18000000 nodes, increment 200000
# 10 iterations of a given list length
# idiomatic list in-memory ordering
# math/rand random numbers as list node values
# nodes 24 bytes in size
# randomly chosen data data values
1000    8715    8727    8727
201000  3290622 3349706 3349706
...
```

Four columns of output:

1. List length in nodes
2. Mean count of comparisons, 10 iterations on the list length, recursive algorithm
3. Mean count of comparisons, 10 iterations on the list length, wikipedia bottom up algorithm
4. Mean count of comparisons, 10 iterations on the list length, July 2021 iterative algorithm

After each sort, the list data is reset,
so randomly-chosen data value lists are the same for each algorithm.
It's not strictly necessary to do 10 iterations on presorted data lists.

You can use

* randomly-chosen data, which can cause more or less comparisons in a sort
* presorted data
* presorted, reverse ordered, data

Presorted data (both kinds) should have the same number of comparisons every time.
