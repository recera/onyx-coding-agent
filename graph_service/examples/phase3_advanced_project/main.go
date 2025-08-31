package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Generic interfaces and types (Go 1.18+ features)
type Comparable[T any] interface {
	Compare(other T) int
}

type Processor[T any] interface {
	Process(item T) T
}

type Cache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]V),
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.items[key]
	return value, exists
}

// Worker Pool Pattern Implementation
type Job struct {
	ID     int
	Data   string
	Result chan<- string
}

type WorkerPool struct {
	jobs    chan Job
	workers int
	wg      sync.WaitGroup
}

func NewWorkerPool(numWorkers int, jobQueueSize int) *WorkerPool {
	return &WorkerPool{
		jobs:    make(chan Job, jobQueueSize),
		workers: numWorkers,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()

	for {
		select {
		case job := <-wp.jobs:
			// Process the job
			result := fmt.Sprintf("Worker %d processed job %d: %s", id, job.ID, job.Data)

			// Send result back
			select {
			case job.Result <- result:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			fmt.Printf("Worker %d shutting down\n", id)
			return
		}
	}
}

func (wp *WorkerPool) Submit(job Job) {
	wp.jobs <- job
}

func (wp *WorkerPool) Close() {
	close(wp.jobs)
	wp.wg.Wait()
}

// Pipeline Pattern Implementation
type Pipeline[T any] struct {
	stages []PipelineStage[T]
}

type PipelineStage[T any] func(<-chan T) <-chan T

func NewPipeline[T any](stages ...PipelineStage[T]) *Pipeline[T] {
	return &Pipeline[T]{stages: stages}
}

func (p *Pipeline[T]) Execute(input <-chan T) <-chan T {
	current := input
	for _, stage := range p.stages {
		current = stage(current)
	}
	return current
}

// Producer-Consumer Pattern with Generics
type Producer[T any] struct {
	output chan<- T
}

func NewProducer[T any](output chan<- T) *Producer[T] {
	return &Producer[T]{output: output}
}

func (p *Producer[T]) Produce(ctx context.Context, items ...T) {
	go func() {
		defer close(p.output)
		for _, item := range items {
			select {
			case p.output <- item:
			case <-ctx.Done():
				return
			}
		}
	}()
}

type Consumer[T any] struct {
	input <-chan T
}

func NewConsumer[T any](input <-chan T) *Consumer[T] {
	return &Consumer[T]{input: input}
}

func (c *Consumer[T]) Consume(ctx context.Context, processor func(T)) {
	go func() {
		for {
			select {
			case item, ok := <-c.input:
				if !ok {
					return
				}
				processor(item)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Fan-out/Fan-in Pattern
type FanOut[T any] struct {
	input   <-chan T
	outputs []chan T
}

func NewFanOut[T any](input <-chan T, numOutputs int) *FanOut[T] {
	outputs := make([]chan T, numOutputs)
	for i := 0; i < numOutputs; i++ {
		outputs[i] = make(chan T)
	}

	return &FanOut[T]{
		input:   input,
		outputs: outputs,
	}
}

func (fo *FanOut[T]) Start(ctx context.Context) {
	go func() {
		defer func() {
			for _, output := range fo.outputs {
				close(output)
			}
		}()

		i := 0
		for {
			select {
			case item, ok := <-fo.input:
				if !ok {
					return
				}
				// Round-robin distribution
				select {
				case fo.outputs[i%len(fo.outputs)] <- item:
					i++
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (fo *FanOut[T]) GetOutputs() []<-chan T {
	outputs := make([]<-chan T, len(fo.outputs))
	for i, output := range fo.outputs {
		outputs[i] = output
	}
	return outputs
}

// Fan-in Pattern
func FanIn[T any](ctx context.Context, inputs ...<-chan T) <-chan T {
	output := make(chan T)
	var wg sync.WaitGroup

	for _, input := range inputs {
		wg.Add(1)
		go func(ch <-chan T) {
			defer wg.Done()
			for {
				select {
				case item, ok := <-ch:
					if !ok {
						return
					}
					select {
					case output <- item:
					case <-ctx.Done():
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}(input)
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

// HTTP API Integration (would call Python services in cross-language setup)
type APIClient struct {
	baseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{baseURL: baseURL}
}

func (api *APIClient) CallPythonService(endpoint string, data interface{}) error {
	// In a real implementation, this would make HTTP calls to Python services
	// This demonstrates cross-language communication patterns
	fmt.Printf("Calling Python service: %s%s with data: %v\n", api.baseURL, endpoint, data)
	return nil
}

// Main demonstration function
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("=== Phase 3: Advanced Go Features Demo ===")

	// Demonstrate Generic Cache
	fmt.Println("\n1. Generic Cache Demo:")
	stringCache := NewCache[string, int]()
	stringCache.Set("hello", 42)
	stringCache.Set("world", 100)

	if value, exists := stringCache.Get("hello"); exists {
		fmt.Printf("Cache value for 'hello': %d\n", value)
	}

	// Demonstrate Worker Pool Pattern
	fmt.Println("\n2. Worker Pool Demo:")
	pool := NewWorkerPool(3, 10)
	pool.Start(ctx)

	// Submit some jobs
	for i := 0; i < 5; i++ {
		resultChan := make(chan string, 1)
		job := Job{
			ID:     i,
			Data:   fmt.Sprintf("Task %d", i),
			Result: resultChan,
		}

		go func(j Job, resCh <-chan string) {
			pool.Submit(j)
			if result := <-resCh; result != "" {
				fmt.Printf("Job result: %s\n", result)
			}
		}(job, resultChan)
	}

	// Demonstrate Pipeline Pattern
	fmt.Println("\n3. Pipeline Demo:")
	input := make(chan int, 5)

	// Pipeline stages
	double := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for x := range in {
				out <- x * 2
			}
		}()
		return out
	}

	addTen := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for x := range in {
				out <- x + 10
			}
		}()
		return out
	}

	pipeline := NewPipeline(double, addTen)
	output := pipeline.Execute(input)

	// Send data through pipeline
	go func() {
		defer close(input)
		for i := 1; i <= 3; i++ {
			input <- i
		}
	}()

	// Read results
	for result := range output {
		fmt.Printf("Pipeline result: %d\n", result)
	}

	// Demonstrate Producer-Consumer Pattern
	fmt.Println("\n4. Producer-Consumer Demo:")
	channel := make(chan string, 5)
	producer := NewProducer(channel)
	consumer := NewConsumer(channel)

	// Start consumer
	consumer.Consume(ctx, func(item string) {
		fmt.Printf("Consumed: %s\n", item)
	})

	// Start producer
	producer.Produce(ctx, "item1", "item2", "item3")

	// Demonstrate Cross-Language API calls
	fmt.Println("\n5. Cross-Language API Demo:")
	apiClient := NewAPIClient("http://localhost:8000")
	err := apiClient.CallPythonService("/api/process", map[string]string{
		"action": "analyze",
		"data":   "golang_code_sample",
	})
	if err != nil {
		log.Printf("API call failed: %v", err)
	}

	// Wait a bit for async operations to complete
	time.Sleep(2 * time.Second)

	// Clean up
	pool.Close()

	fmt.Println("\n=== Demo Complete ===")
}
