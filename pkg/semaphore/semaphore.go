// Joseph Bursey <jbursey@tevora.com>

package semaphore

type semchan chan struct{}

type Semaphore struct {
    sem semchan
}

func New(max int) *Semaphore {
    return &Semaphore{
        sem: make(chan struct{}, max),
    }
}

func (s *Semaphore) Acquire() {
    s.sem <- struct{}{}
}

func (s *Semaphore) Release() {
    <- s.sem
}
