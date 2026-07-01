// Joseph Bursey <jbursey@tevora.com>

// I totally didn't steal this from https://pkg.go.dev/container/heap
// This implementation is specifically for Job

package job

type PriorityQueue []*Job

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	curJob := x.(*Job)
	curJob.index = len(*pq)
	*pq = append(*pq, curJob)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	curJob := old[n - 1]
	old[n - 1] = nil
	curJob.index = -1
	*pq = old[0 : n - 1]
	return curJob
}
