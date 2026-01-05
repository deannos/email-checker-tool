package worker

import "sync"

func Run[T any, R any](workers int, jobs []T, fn func(T) R) []R {
	jobCh := make(chan T)
	resCh := make(chan R)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobCh {
				resCh <- fn(j)
			}
		}()
	}

	go func() {
		for _, j := range jobs {
			jobCh <- j
		}
		close(jobCh)
	}()

	go func() {
		wg.Wait()
		close(resCh)
	}()

	var results []R
	for r := range resCh {
		results = append(results, r)
	}
	return results
}
