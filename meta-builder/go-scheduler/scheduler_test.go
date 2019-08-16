package go_scheduler

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testSchedulerTask(ctx context.Context) error {
	return nil
}

func TestNewScheduler(t *testing.T) {
	scheduler := New()

	assert.Zero(t, scheduler.TaskCount(), "scheduler.TaskCount() should equal 0")
	assert.Empty(t, scheduler.Tasks(), "scheduler.Tasks() should equal nil")
	assert.Equal(t, runtime.NumCPU(), scheduler.options.ConcurrentTasks,
		"scheduler.options.ConcurrentTasks should equal runtime.NumCPU()")

	scheduler = New(ConcurrentTasks(3))

	assert.Zero(t, scheduler.TaskCount(), "scheduler.TaskCount() should equal 0")
	assert.Empty(t, scheduler.Tasks(), "scheduler.Tasks() should equal nil")
	assert.Equal(t, 3, scheduler.options.ConcurrentTasks, "scheduler.options.ConcurrentTasks should equal 3")
}

func TestSchedulerAddAndRemoveTask(t *testing.T) {
	scheduler := New()
	a := scheduler.AddTask(testSchedulerTask)
	b := scheduler.AddTask(testSchedulerTask)
	c := scheduler.AddTask(testSchedulerTask)

	assert.Equal(t, 3, scheduler.TaskCount(), "scheduler.TaskCount() should equal 3")
	assert.Equal(t, []*Task{a, b, c}, scheduler.Tasks(), "scheduler.Tasks() should equal [a, b, c]")

	scheduler.RemoveTask(a)
	scheduler.RemoveTask(c)

	assert.Equal(t, 1, scheduler.TaskCount(), "scheduler.TaskCount() should equal 1")
	assert.Equal(t, []*Task{b}, scheduler.Tasks(), "scheduler.Tasks() should equal [b]")
}

func TestSchedulerRemoveTasks(t *testing.T) {
	scheduler := New()
	a := scheduler.AddTask(testSchedulerTask)
	b := scheduler.AddTask(testSchedulerTask)
	c := scheduler.AddTask(testSchedulerTask)

	scheduler.RemoveTasks(a, c)

	assert.Equal(t, 1, scheduler.TaskCount(), "scheduler.TaskCount() should equal 1")
	assert.Equal(t, []*Task{b}, scheduler.Tasks(), "scheduler.Tasks() should equal [b]")
}

func TestSchedulerAddAndRemoveDependency(t *testing.T) {
	scheduler := New()
	a := scheduler.AddTask(testSchedulerTask)
	b := scheduler.AddTask(testSchedulerTask)
	c := scheduler.AddTask(testSchedulerTask)
	d := scheduler.AddTask(testSchedulerTask)

	scheduler.AddDependency(a, b)
	scheduler.AddDependency(b, c)
	scheduler.AddDependency(c, d)

	assert.Equal(t, 1, scheduler.DependencyCount(a), " scheduler.DependencyCount(a) should equal 1")
	assert.Equal(t, 1, scheduler.DependencyCount(b), " scheduler.DependencyCount(b) should equal 1")
	assert.Equal(t, 1, scheduler.DependencyCount(c), " scheduler.DependencyCount(c) should equal 1")
	assert.Equal(t, 0, scheduler.DependencyCount(d), " scheduler.DependencyCount(d) should equal 0")

	assert.Equal(t, []*Task{b}, scheduler.Dependencies(a), " scheduler.Dependencies(a) should equal [b]")
	assert.Equal(t, []*Task{c}, scheduler.Dependencies(b), " scheduler.Dependencies(b) should equal [c]")
	assert.Equal(t, []*Task{d}, scheduler.Dependencies(c), " scheduler.Dependencies(c) should equal [d]")
	assert.Empty(t, scheduler.Dependencies(d), " scheduler.Dependencies(d) should be empty")

	scheduler.RemoveDependency(a, b)
	scheduler.RemoveDependency(c, d)

	assert.Equal(t, 0, scheduler.DependencyCount(a), " scheduler.DependencyCount(a) should equal 0")
	assert.Equal(t, 1, scheduler.DependencyCount(b), " scheduler.DependencyCount(b) should equal 1")
	assert.Equal(t, 0, scheduler.DependencyCount(c), " scheduler.DependencyCount(c) should equal 0")
	assert.Equal(t, 0, scheduler.DependencyCount(d), " scheduler.DependencyCount(d) should equal 0")

	assert.Empty(t, scheduler.Dependencies(a), " scheduler.Dependencies(a) should be empty")
	assert.Equal(t, []*Task{c}, scheduler.Dependencies(b), " scheduler.Dependencies(b) should be empty")
	assert.Empty(t, scheduler.Dependencies(c), " scheduler.Dependencies(c) should equal [d]")
	assert.Empty(t, scheduler.Dependencies(d), " scheduler.Dependencies(d) should be empty")
}

func TestSchedulerRun(t *testing.T) {
	state := sync.Map{}
	state.Store("a", false)
	state.Store("b", false)
	state.Store("c", false)
	state.Store("d", false)
	state.Store("e", false)
	state.Store("f", false)

	checkState := func(key string, expected map[string]bool) func(context.Context) error {
		return func(ctx context.Context) error {
			for key, value := range expected {
				actual, ok := state.Load(key)
				assert.True(t, ok, "state["+key+"] should exist")

				if value {
					assert.True(t, actual.(bool), "state["+key+"] should equal true")
				} else {
					assert.False(t, actual.(bool), "state["+key+"] should equal false")
				}
			}

			time.Sleep(100 * time.Millisecond)
			state.Store(key, true)

			return nil
		}
	}

	scheduler := New()

	a := scheduler.AddTask(checkState("a", map[string]bool{
		"a": false,
		"b": false,
		"c": false,
		"d": false,
		"e": false,
		"f": false,
	}))

	b := scheduler.AddTask(checkState("b", map[string]bool{
		"a": true,
		"b": false,
		"c": false,
		"d": false,
		"e": false,
		"f": false,
	}))

	c := scheduler.AddTask(checkState("c", map[string]bool{
		"a": true,
		"b": true,
		"c": false,
		"d": false,
		"e": false,
		"f": false,
	}))

	d := scheduler.AddTask(checkState("d", map[string]bool{
		"a": true,
		"b": true,
		"c": true,
		"d": false,
		"e": false,
		"f": false,
	}))

	e := scheduler.AddTask(checkState("e", map[string]bool{
		"a": true,
		"b": true,
		"c": true,
		"d": false,
		"e": false,
		"f": false,
	}))

	f := scheduler.AddTask(checkState("f", map[string]bool{
		"a": true,
		"b": true,
		"c": true,
		"d": true,
		"e": true,
		"f": false,
	}))

	scheduler.AddDependency(f, e)
	scheduler.AddDependency(f, d)
	scheduler.AddDependency(e, c)
	scheduler.AddDependency(d, c)
	scheduler.AddDependency(c, b)
	scheduler.AddDependency(b, a)

	err := scheduler.Run(context.Background())
	assert.NoError(t, err, "scheduler.Run() should not return an error")

	actual, ok := state.Load("f")
	assert.True(t, ok, "state[f] should exist")
	assert.True(t, actual.(bool), "state[f] should equal true")
}

func TestSchedulerRunCircular(t *testing.T) {
	scheduler := New()

	a := scheduler.AddTask(testSchedulerTask)
	b := scheduler.AddTask(testSchedulerTask)

	scheduler.AddDependency(a, b)
	scheduler.AddDependency(b, a)

	err := scheduler.Run(context.Background())
	assert.EqualError(t, ErrCircularDependency, err.Error(), "scheduler.Run() should return ErrCircularDependency")
}

func TestSchedulerRunTaskError(t *testing.T) {
	state := sync.Map{}
	state.Store("a", false)
	state.Store("b", false)

	scheduler := New()
	testError := errors.New("an error")

	a := scheduler.AddTask(func(ctx context.Context) error {
		state.Store("a", true)
		return testError
	})

	b := scheduler.AddTask(func(ctx context.Context) error {
		state.Store("b", true)
		return nil
	})

	scheduler.AddDependency(b, a)

	err := scheduler.Run(context.Background())
	assert.EqualError(t, testError, err.Error(), "scheduler.Run() should return testError")

	actual, ok := state.Load("a")
	assert.True(t, ok, "state[a] should exist")
	assert.True(t, actual.(bool), "state[a] should equal true")

	actual, ok = state.Load("b")
	assert.True(t, ok, "state[b] should exist")
	assert.False(t, actual.(bool), "state[b] should equal false")
}

func TestSchedulerRunResizeLevels(t *testing.T) {
	scheduler := New()
	a := scheduler.AddTask(testSchedulerTask)
	b := scheduler.AddTask(testSchedulerTask)

	scheduler.AddDependency(b, a)

	err := scheduler.Run(context.Background())
	assert.NoError(t, err, "scheduler.Run() should not return an error")

	// remove the dependency to reduce the number of concurrency levels from 2 to 1
	scheduler.RemoveDependency(b, a)

	err = scheduler.Run(context.Background())
	assert.NoError(t, err, "scheduler.Run() should not return an error")
}

func TestSchedulerRunResizeTasks(t *testing.T) {
	scheduler := New()
	a := scheduler.AddTask(testSchedulerTask)
	b := scheduler.AddTask(testSchedulerTask)
	c := scheduler.AddTask(testSchedulerTask)

	scheduler.AddDependency(b, a)
	scheduler.AddDependency(c, a)

	err := scheduler.Run(context.Background())
	assert.NoError(t, err, "scheduler.Run() should not return an error")

	// remove the dependency to reduce the number of tasks in the 2nd concurrency level from 2 to 1
	scheduler.RemoveDependency(c, a)

	err = scheduler.Run(context.Background())
	assert.NoError(t, err, "scheduler.Run() should not return an error")
}
