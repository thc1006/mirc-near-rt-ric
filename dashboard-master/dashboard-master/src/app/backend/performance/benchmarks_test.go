// Copyright 2017 The Kubernetes Authors.
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

package performance

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kubernetes/dashboard/src/app/backend/cache"
	"github.com/kubernetes/dashboard/src/app/backend/informer"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/optimized"
	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"
)

// BenchmarkOriginalPodList benchmarks the original pod list implementation
func BenchmarkOriginalPodList(b *testing.B) {
	client := fake.NewSimpleClientset()
	pods := generateTestPods(1000, 10) // 1000 pods across 10 namespaces
	
	// Add pods to fake client
	for _, testPod := range pods {
		client.CoreV1().Pods(testPod.Namespace).Create(context.TODO(), &testPod, metav1.CreateOptions{})
	}
	
	nsQuery := common.NewNamespaceQuery([]string{})
	dsQuery := &dataselect.DataSelectQuery{
		PaginationQuery: dataselect.NoPagination,
		SortQuery:       dataselect.NoSort,
		FilterQuery:     dataselect.NoFilter,
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := pod.GetPodList(client, nil, nsQuery, dsQuery)
		if err != nil {
			b.Fatalf("Error getting pod list: %v", err)
		}
	}
}

// BenchmarkOptimizedPodList benchmarks the optimized pod list implementation
func BenchmarkOptimizedPodList(b *testing.B) {
	client := fake.NewSimpleClientset()
	pods := generateTestPods(1000, 10)
	
	// Add pods to fake client
	for _, testPod := range pods {
		client.CoreV1().Pods(testPod.Namespace).Create(context.TODO(), &testPod, metav1.CreateOptions{})
	}
	
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
	
	// Setup optimized components
	resourceCache := cache.NewResourceCache(ctx, client, logger)
	resourceCache.Start()
	defer resourceCache.Stop()
	
	optimizedService := optimized.NewOptimizedPodListService(client, resourceCache, logger)
	
	nsQuery := common.NewNamespaceQuery([]string{})
	dsQuery := &dataselect.DataSelectQuery{
		PaginationQuery: dataselect.NoPagination,
		SortQuery:       dataselect.NoSort,
		FilterQuery:     dataselect.NoFilter,
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := optimizedService.GetPodList(ctx, nsQuery, dsQuery, nil)
		if err != nil {
			b.Fatalf("Error getting optimized pod list: %v", err)
		}
	}
}

// BenchmarkInformerPodList benchmarks the informer-based pod list implementation
func BenchmarkInformerPodList(b *testing.B) {
	client := fake.NewSimpleClientset()
	pods := generateTestPods(1000, 10)
	
	// Add pods to fake client
	for _, testPod := range pods {
		client.CoreV1().Pods(testPod.Namespace).Create(context.TODO(), &testPod, metav1.CreateOptions{})
	}
	
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
	
	// Setup informer manager
	informerMgr := informer.NewInformerManager(ctx, client, logger)
	err := informerMgr.Start()
	if err != nil {
		b.Fatalf("Failed to start informer manager: %v", err)
	}
	defer informerMgr.Stop()
	
	// Wait for cache sync
	time.Sleep(100 * time.Millisecond)
	
	nsQuery := common.NewNamespaceQuery([]string{})
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := informerMgr.GetPodsFromCache(nsQuery)
		if err != nil {
			b.Fatalf("Error getting pods from informer cache: %v", err)
		}
	}
}

// BenchmarkConcurrentPodListOriginal benchmarks concurrent access to original pod list
func BenchmarkConcurrentPodListOriginal(b *testing.B) {
	client := fake.NewSimpleClientset()
	pods := generateTestPods(500, 5)
	
	for _, testPod := range pods {
		client.CoreV1().Pods(testPod.Namespace).Create(context.TODO(), &testPod, metav1.CreateOptions{})
	}
	
	nsQuery := common.NewNamespaceQuery([]string{})
	dsQuery := &dataselect.DataSelectQuery{
		PaginationQuery: dataselect.NoPagination,
		SortQuery:       dataselect.NoSort,
		FilterQuery:     dataselect.NoFilter,
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := pod.GetPodList(client, nil, nsQuery, dsQuery)
			if err != nil {
				b.Fatalf("Error getting pod list: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentPodListOptimized benchmarks concurrent access to optimized pod list
func BenchmarkConcurrentPodListOptimized(b *testing.B) {
	client := fake.NewSimpleClientset()
	pods := generateTestPods(500, 5)
	
	for _, testPod := range pods {
		client.CoreV1().Pods(testPod.Namespace).Create(context.TODO(), &testPod, metav1.CreateOptions{})
	}
	
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
	
	resourceCache := cache.NewResourceCache(ctx, client, logger)
	resourceCache.Start()
	defer resourceCache.Stop()
	
	optimizedService := optimized.NewOptimizedPodListService(client, resourceCache, logger)
	
	nsQuery := common.NewNamespaceQuery([]string{})
	dsQuery := &dataselect.DataSelectQuery{
		PaginationQuery: dataselect.NoPagination,
		SortQuery:       dataselect.NoSort,
		FilterQuery:     dataselect.NoFilter,
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := optimizedService.GetPodList(ctx, nsQuery, dsQuery, nil)
			if err != nil {
				b.Fatalf("Error getting optimized pod list: %v", err)
			}
		}
	})
}

// BenchmarkCachePerformance benchmarks cache hit/miss performance
func BenchmarkCachePerformance(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
	client := fake.NewSimpleClientset()
	
	resourceCache := cache.NewResourceCache(ctx, client, logger)
	resourceCache.Start()
	defer resourceCache.Stop()
	
	// Pre-populate cache
	testData := generateTestPodList(100)
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("namespace-%d", i)
		resourceCache.SetPodList(key, "", testData)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.Run("CacheHit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("namespace-%d", i%50)
			result := resourceCache.GetPodList(key, "")
			if result == nil {
				b.Fatalf("Expected cache hit but got miss for key: %s", key)
			}
		}
	})
	
	b.Run("CacheMiss", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("missing-namespace-%d", i)
			result := resourceCache.GetPodList(key, "")
			if result != nil {
				b.Fatalf("Expected cache miss but got hit for key: %s", key)
			}
		}
	})
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("OriginalAllocation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Simulate original allocation pattern
			pods := make([]pod.Pod, 1000)
			for j := range pods {
				pods[j] = pod.Pod{
					ObjectMeta:      generateObjectMeta(fmt.Sprintf("pod-%d", j)),
					ContainerImages: make([]string, 5),
					Warnings:        make([]common.Event, 0),
				}
			}
			_ = pods
		}
	})
	
	b.Run("OptimizedAllocation", func(b *testing.B) {
		pool := sync.Pool{
			New: func() interface{} {
				return make([]pod.Pod, 0, 1000)
			},
		}
		
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Simulate optimized allocation pattern
			pods := pool.Get().([]pod.Pod)
			pods = pods[:0] // Reset length but keep capacity
			
			for j := 0; j < 1000; j++ {
				podItem := pod.Pod{
					ObjectMeta:      generateObjectMeta(fmt.Sprintf("pod-%d", j)),
					ContainerImages: make([]string, 0, 5), // Pre-allocate capacity
					Warnings:        make([]common.Event, 0, 2),
				}
				pods = append(pods, podItem)
			}
			
			pool.Put(pods)
		}
	})
}

// BenchmarkDataTransformation benchmarks data transformation operations
func BenchmarkDataTransformation(b *testing.B) {
	pods := generateTestPods(1000, 10)
	
	b.Run("OriginalTransformation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Simulate original transformation
			result := make([]pod.Pod, len(pods))
			for j, k8sPod := range pods {
				result[j] = pod.Pod{
					ObjectMeta:      generateObjectMeta(k8sPod.Name),
					Status:          string(k8sPod.Status.Phase),
					NodeName:        k8sPod.Spec.NodeName,
					ContainerImages: extractContainerImages(k8sPod.Spec.Containers),
				}
			}
			_ = result
		}
	})
	
	b.Run("OptimizedTransformation", func(b *testing.B) {
		stringPool := sync.Pool{
			New: func() interface{} {
				return make([]string, 0, 10)
			},
		}
		
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Simulate optimized transformation
			result := make([]pod.Pod, 0, len(pods))
			for _, k8sPod := range pods {
				images := stringPool.Get().([]string)
				images = images[:0]
				
				for _, container := range k8sPod.Spec.Containers {
					images = append(images, container.Image)
				}
				
				// Copy images since we're returning the slice to pool
				containerImages := make([]string, len(images))
				copy(containerImages, images)
				stringPool.Put(images)
				
				result = append(result, pod.Pod{
					ObjectMeta:      generateObjectMeta(k8sPod.Name),
					Status:          string(k8sPod.Status.Phase),
					NodeName:        k8sPod.Spec.NodeName,
					ContainerImages: containerImages,
				})
			}
			_ = result
		}
	})
}

// Helper functions for generating test data

func generateTestPods(count int, namespaceCount int) []v1.Pod {
	pods := make([]v1.Pod, count)
	namespaces := make([]string, namespaceCount)
	
	for i := 0; i < namespaceCount; i++ {
		namespaces[i] = fmt.Sprintf("namespace-%d", i)
	}
	
	for i := 0; i < count; i++ {
		pods[i] = v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            fmt.Sprintf("pod-%d", i),
				Namespace:       namespaces[i%namespaceCount],
				ResourceVersion: fmt.Sprintf("%d", rand.Intn(1000000)),
				Labels: map[string]string{
					"app":        fmt.Sprintf("app-%d", i%10),
					"version":    fmt.Sprintf("v%d", i%3+1),
					"component":  fmt.Sprintf("comp-%d", i%5),
				},
			},
			Spec: v1.PodSpec{
				NodeName: fmt.Sprintf("node-%d", i%20),
				Containers: []v1.Container{
					{
						Name:  "main",
						Image: fmt.Sprintf("image-%d:latest", i%50),
					},
					{
						Name:  "sidecar",
						Image: fmt.Sprintf("sidecar-%d:v1.0", i%20),
					},
				},
			},
			Status: v1.PodStatus{
				Phase: []v1.PodPhase{v1.PodRunning, v1.PodPending, v1.PodSucceeded}[i%3],
			},
		}
	}
	
	return pods
}

func generateTestPodList(count int) *v1.PodList {
	return &v1.PodList{
		Items: generateTestPods(count, 5),
	}
}

func generateObjectMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:            name,
		Namespace:       "default",
		ResourceVersion: fmt.Sprintf("%d", rand.Intn(1000000)),
	}
}

func extractContainerImages(containers []v1.Container) []string {
	images := make([]string, len(containers))
	for i, container := range containers {
		images[i] = container.Image
	}
	return images
}

// Load testing helpers

// LoadTest runs a sustained load test
func LoadTest(b *testing.B, name string, testFunc func() error) {
	b.Run(name, func(b *testing.B) {
		var wg sync.WaitGroup
		concurrency := 10
		duration := 10 * time.Second
		
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		defer cancel()
		
		b.ResetTimer()
		start := time.Now()
		
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				count := 0
				for {
					select {
					case <-ctx.Done():
						return
					default:
						if err := testFunc(); err != nil {
							b.Errorf("Load test error: %v", err)
							return
						}
						count++
					}
				}
			}()
		}
		
		wg.Wait()
		elapsed := time.Since(start)
		
		b.Logf("Load test completed in %v", elapsed)
	})
}

// Example load tests
func BenchmarkLoadTestOriginal(b *testing.B) {
	client := fake.NewSimpleClientset()
	pods := generateTestPods(100, 5)
	
	for _, testPod := range pods {
		client.CoreV1().Pods(testPod.Namespace).Create(context.TODO(), &testPod, metav1.CreateOptions{})
	}
	
	nsQuery := common.NewNamespaceQuery([]string{})
	dsQuery := &dataselect.DataSelectQuery{
		PaginationQuery: dataselect.NoPagination,
		SortQuery:       dataselect.NoSort,
		FilterQuery:     dataselect.NoFilter,
	}
	
	LoadTest(b, "OriginalPodList", func() error {
		_, err := pod.GetPodList(client, nil, nsQuery, dsQuery)
		return err
	})
}

func BenchmarkLoadTestOptimized(b *testing.B) {
	client := fake.NewSimpleClientset()
	pods := generateTestPods(100, 5)
	
	for _, testPod := range pods {
		client.CoreV1().Pods(testPod.Namespace).Create(context.TODO(), &testPod, metav1.CreateOptions{})
	}
	
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
	
	resourceCache := cache.NewResourceCache(ctx, client, logger)
	resourceCache.Start()
	defer resourceCache.Stop()
	
	optimizedService := optimized.NewOptimizedPodListService(client, resourceCache, logger)
	
	nsQuery := common.NewNamespaceQuery([]string{})
	dsQuery := &dataselect.DataSelectQuery{
		PaginationQuery: dataselect.NoPagination,
		SortQuery:       dataselect.NoSort,
		FilterQuery:     dataselect.NoFilter,
	}
	
	LoadTest(b, "OptimizedPodList", func() error {
		_, err := optimizedService.GetPodList(ctx, nsQuery, dsQuery, nil)
		return err
	})
}