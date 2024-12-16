#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/time.h>
#include <string.h>
#include <sys/utsname.h>
#include <time.h>

float elapsed_time(struct timeval before, struct timeval after);
char *hostname(void);
char *time_stamp(void);

struct Node {
	int Data;
	struct Node *Next;
};


struct Node *freeNodeList = NULL;
int malloced_nodes = 0;
int free_list_count = 0;
struct Node *new_node(int value, struct Node *next);
void free_list(struct Node *head);

int preallocated = 0;
struct Node *arrayallocation = NULL;
int preallocated_node_count = 0;
int next_node_index;
void perform_preallocation(int node_count);

struct Node *mergesort(struct Node *head);
struct Node *randomValueList(int n);
int isSorted(struct Node *head);

int
main(int ac, char **av)
{
	int c, n;
	float total = 0.0;
	struct timeval tv;
	int count_begin = 1000;
	int count_until = 18000000;
	int increment = 200000;
	char *hn = hostname();

	gettimeofday(&tv, NULL);
	srandom(tv.tv_usec | getpid());

    while (EOF != (c = getopt(ac, av, "b:i:pu:")))
    {
        switch (c)
        {
		case 'b':
			count_begin = atoi(optarg);
			break;
		case 'i':
			increment = atoi(optarg);
			break;
        case 'p':
			preallocated = 1;
			break;
        case 'u':
			count_until = atoi(optarg);
			break;
		}
	}

	printf("# %s on %s\n", time_stamp(), hn);
	printf("# Start at %d nodes, end before %d nodes, increment %d\n",
		count_begin, count_until, increment);

	for (n = count_begin; n < count_until; n += increment) {
		int i;

		float max = -1.0;
		float min = 1.0E9;

		if (preallocated) {
			perform_preallocation(n);
		}

		for (i = 0; i < 11; i++) {
			struct timeval before, after;
			float et;

			struct Node *nl, *head = randomValueList(n);

			gettimeofday(&before, NULL);
			nl = mergesort(head);
			gettimeofday(&after, NULL);

			if (i > 0) {
				if (!isSorted(nl)) {
					fprintf(stderr, "n %d, i %d, final list not sorted\n", n, i);
				}
				/* Don't time first sort, warm the cache */
				et = elapsed_time(before, after);
				if (et > max) { max = et; }
				if (et < min) { min = et; }
				total += et;
			}

			free_list(nl);
		}

		total /= 10.;
		printf("%d\t%.04f\t%.04f\t%.04f\n", n, total, min, max);
		fflush(stdout);
	}

	printf("# ending at %s on %s\n", time_stamp(), hn);

	return 0;
}

int
isSorted(struct Node *head) {
	if (head == NULL || head->Next == NULL) {
		return 1;
	}
	for (; head->Next != NULL; head = head->Next) {
		if (head->Data > head->Next->Data) {
			return 0;
		}
	}
	return 1;
}

struct Node *
randomValueList(int n) {
	struct Node *head = NULL;
	int i;
	for (i = 0; i < n; i++) {
		head = new_node((int)random(), head);
	}
	return head;
}

void append(struct Node *n, struct Node **hd, struct Node **tl) {
	if (!*hd) {
		*hd = n;
		*tl = n;
		return;
	}
	(*tl)->Next = n;
	*tl = n;
}

struct Node *
mergesort(struct Node *head) {
	struct Node *hd, *tl, *p;
	int mergecount, k;

	hd = tl = NULL;
	p = head;
	mergecount = 2; /* just to pass the first for-test */

	for (k = 1; mergecount > 1; k *= 2) {

		mergecount = 0;

		while (p) {
			int i, psize = 0, qsize;
			struct Node *q = p;

			for (i = 0; q && i < k; i++) {
				psize++;
				q = q->Next;
			}

			qsize = psize;

			while (psize > 0 && qsize > 0 && q) {
				if (p->Data < q->Data) {
					append(p, &hd, &tl);
					p = p->Next;
					psize--;
					continue;
				}
				append(q, &hd, &tl);
				q = q->Next;
				qsize--;
			}

			for (; psize > 0 && p; psize--) {
				append(p, &hd, &tl);
				p = p->Next;
			}

			for (; qsize > 0 && q; qsize--) {
				append(q, &hd, &tl);
				q = q->Next;
			}

			p = q;

			mergecount++;
		}

		p = hd;
		head = hd;

		hd = NULL;
		tl->Next = NULL;
		tl = NULL;
	}

	return head;
}

struct Node *
new_node(int value, struct Node *next) {
	if (preallocated) {
		struct Node *p = NULL;
		if (next_node_index < preallocated_node_count) {
			p = &(arrayallocation[next_node_index++]);
			p->Data = value;
			p->Next = next;
		}
		return p;
	}
	struct Node *n;
	if (freeNodeList) {
		n = freeNodeList;
		freeNodeList = freeNodeList->Next;
		--free_list_count;
	} else {
		n = malloc(sizeof(*n));
		malloced_nodes++;
	}
	n->Data = value;
	n->Next = next;
	return n;
}

/* frees entire sorted list at end of benchmarking run */
void
free_list(struct Node *head) {
	if (head == NULL)
		return;
	if (preallocated) {
		memset(arrayallocation, 0, sizeof(struct Node)*preallocated_node_count);
		next_node_index = 0;
		return;
	}
	while (head != NULL) {
		struct Node *tmp = head->Next;
		head->Data = -1;
		head->Next = freeNodeList;
		freeNodeList = head;
		++free_list_count;
		head = tmp;
	}
}

void
really_free_list(struct Node *head) {
	struct Node *tmp;
	if (head == NULL)
		return;
	for (tmp = head; head != NULL; head = tmp->Next) {
		head->Next = freeNodeList;
		freeNodeList = head;
		--free_list_count;
	}
}

/* utility function elapsed_time() */
float
elapsed_time(struct timeval before, struct timeval after)
{
    float r = 0.0;

    if (before.tv_usec > after.tv_usec)
    {
        after.tv_usec += 1000000;
        --after.tv_sec;
    }

    r = (float)(after.tv_sec - before.tv_sec)
        + (1.0E-6)*(float)(after.tv_usec - before.tv_usec);

    return r;
}

void
perform_preallocation(int node_count)
{
	if (!preallocated)
		return;
	if (preallocated_node_count >= node_count)
		return;

	if (arrayallocation) {
		free(arrayallocation);
		arrayallocation = NULL;
	}
	preallocated_node_count = node_count;
	next_node_index = 0;

	arrayallocation = calloc(preallocated_node_count, sizeof(struct Node));
}

struct utsname ubuf;

char *
hostname(void)
{
	if (uname(&ubuf) < 0) {
		return NULL;
	}
	return ubuf.nodename;
}

char time_buf[256];

char *
time_stamp(void) {
    time_t t;
    struct tm *tmp;

    t = time(NULL);
    tmp = localtime(&t);

    strftime(time_buf, sizeof(time_buf),
        "%FT%H:%M:%S%z",
        tmp
    );

	return time_buf;
}
