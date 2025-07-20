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
	Task      func()
	TrainingJob *TrainingJob
}

// JobQueue manages a queue of jobs to be executed by a pool of workers.
type JobQueue struct {
	jobs      chan Job
	semaphore chan struct{}
	wg        sync.WaitGroup
	mutex     sync.Mutex
}

// NewJobQueue creates a new job queue with a given capacity.
func NewJobQueue(capacity int) *JobQueue {
	return &JobQueue{
		jobs:      make(chan Job),
		semaphore: make(chan struct{}, capacity),
	}
}

// Submit adds a new job to the queue.
func (jq *JobQueue) Submit(trainingJob *TrainingJob, task func()) {
	jq.jobs <- Job{Task: task, TrainingJob: trainingJob}
}

// Start begins processing jobs from the queue.
func (jq *JobQueue) Start(ctx context.Context) {
	jq.wg.Add(1)
	go func() {
		defer jq.wg.Done()
		for {
			select {
			case job := <-jq.jobs:
				jq.semaphore <- struct{}{}
				jq.wg.Add(1)
				go func(job Job) {
					defer func() {
						<-jq.semaphore
						jq.wg.Done()
					}()
					job.Task()
				}(job)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop waits for all jobs to complete.
func (jq *JobQueue) Stop() {
	jq.wg.Wait()
}

// SetCapacity updates the capacity of the job queue.
func (jq *JobQueue) SetCapacity(capacity int) {
	jq.mutex.Lock()
	defer jq.mutex.Unlock()

	newSemaphore := make(chan struct{}, capacity)

	// Copy existing tokens to the new semaphore
	for i := 0; i < len(jq.semaphore); i++ {
		newSemaphore <- struct{}{}
	}

	jq.semaphore = newSemaphore
}
