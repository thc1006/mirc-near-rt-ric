// Copyright 2024 The O-RAN Near-RT RIC Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package federatedlearning

import (
	"context"
	"sync"
)

// Job represents a unit of work to be executed.
type Job struct {
	Task     func()
	Priority int
	Details  *TrainingJob
}

// JobQueue manages a priority queue of jobs to be executed by a pool of workers.
type JobQueue struct {
	jobs        chan Job
	wg          sync.WaitGroup
	maxWorkers  int
	workerCount int
	mu          sync.Mutex
}

// NewJobQueue creates a new job queue with a given number of workers.
func NewJobQueue(maxWorkers int) *JobQueue {
	return &JobQueue{
		jobs:       make(chan Job),
		maxWorkers: maxWorkers,
	}
}

// Start begins processing jobs from the queue.
func (q *JobQueue) Start(ctx context.Context) {
	for i := 0; i < q.maxWorkers; i++ {
		q.wg.Add(1)
		go q.worker(ctx)
	}
	q.wg.Wait()
}

// worker is a goroutine that processes jobs from the queue.
func (q *JobQueue) worker(ctx context.Context) {
	defer q.wg.Done()
	for {
		select {
		case job, ok := <-q.jobs:
			if !ok {
				return
			}
			job.Task()
		case <-ctx.Done():
			return
		}
	}
}

// Submit adds a new job to the queue.
func (q *JobQueue) Submit(jobDetails *TrainingJob, task func()) {
	job := Job{
		Task:     task,
		Priority: int(jobDetails.Spec.Priority),
		Details:  jobDetails,
	}
	q.jobs <- job
}

// Stop gracefully shuts down the job queue and its workers.
func (q *JobQueue) Stop() {
	close(q.jobs)
}
