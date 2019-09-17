package go_scheduler

import (
	"context"
	"errors"
	"runtime"

	dag "github.com/blackducksoftware/synopsys-operator/go-directed-acyclic-graph"
)

// Scheduler related errors.
var (
	ErrCircularDependency = errors.New("A task has a circular dependency")
	ErrOrderFailure       = errors.New("A problem occurred sorting the tasks in the correct order")
)

// Scheduler is a concurrent task scheduler.
type Scheduler struct {
	options *Options
	graph   *dag.DirectedGraph
	dirty   bool
	tasks   [][]*Task
	runner  *taskRunner
}

// New returns a concurrent task scheduler.
// Tasks may be dependent on each, being sorted into a layered topological order
// using the Coffman-Graham algorithm.
func New(options ...Option) *Scheduler {
	scheduler := &Scheduler{
		graph:  dag.NewDirectedGraph(),
		tasks:  make([][]*Task, 0),
		runner: newTaskRunner(),
		options: &Options{
			ConcurrentTasks: runtime.NumCPU(),
		},
	}

	for _, option := range options {
		option(scheduler.options)
	}

	return scheduler
}

// Tasks returns a list of the tasks registered with the scheduler.
func (s *Scheduler) Tasks() []*Task {
	nodes := s.graph.Nodes()
	tasks := make([]*Task, 0, len(nodes))

	for _, node := range nodes {
		if task, ok := node.(*Task); ok {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// TaskCount returns the number of tasks registered with the scheduler.
func (s *Scheduler) TaskCount() int {
	return s.graph.NodeCount()
}

// AddTask registers the function with the scheduler and returns a Task.
func (s *Scheduler) AddTask(function TaskFunc) *Task {
	task := newTask(s, function)
	s.graph.AddNode(task)
	s.invalidate()
	return task
}

// RemoveTask removes the specified task from the scheduler.
func (s *Scheduler) RemoveTask(task *Task) {
	s.graph.RemoveNode(task)
	s.invalidate()
}

// RemoveTasks removes the specified tasks from the scheduler.
func (s *Scheduler) RemoveTasks(tasks ...*Task) {
	nodes := make([]interface{}, len(tasks))
	for i, task := range tasks {
		nodes[i] = task
	}
	s.graph.RemoveNodes(nodes...)
	s.invalidate()
}

// Dependencies lists the tasks the specified task depends on.
func (s *Scheduler) Dependencies(task *Task) []*Task {
	edges := s.graph.IncomingEdges(task)

	dependencies := make([]*Task, len(edges))
	for i, edge := range edges {
		dependencies[i] = edge.(*Task)
	}
	return dependencies
}

// DependencyCount returns the number of tasks the specified task depends on.
func (s *Scheduler) DependencyCount(task *Task) int {
	return s.graph.IncomingEdgeCount(task)
}

// AddDependency creates a dependency between the specified task itself
// and the dependency task.
// When ran, the scheduler ensures the dependency task is executed first.
func (s *Scheduler) AddDependency(task *Task, dependency *Task) {
	s.graph.AddEdge(dependency, task)
	s.invalidate()
}

// RemoveDependency removes the dependency between the specified task itself
// and the dependency task.
func (s *Scheduler) RemoveDependency(task *Task, dependency *Task) {
	s.graph.RemoveEdge(dependency, task)
	s.invalidate()
}

func (s *Scheduler) invalidate() {
	s.dirty = true
}

func (s *Scheduler) resizeLevels(count int) {
	currentCount := len(s.tasks)

	// extend or shrink the task levels slice depending on the difference
	// between the new and old level counts
	if count < currentCount {
		for i := count; i < currentCount; i++ {
			s.tasks[i] = nil
		}

		s.tasks = s.tasks[:count]
	} else if count > currentCount {
		extendBy := count - currentCount
		s.tasks = append(s.tasks, make([][]*Task, extendBy)...)
	}
}

func (s *Scheduler) resizeTasks(level, count int) {
	currentCount := len(s.tasks[level])

	// extend or shrink the tasks slice depending on the difference
	// between the new and old concurrency counts
	if count < currentCount {
		for j := count; j < currentCount; j++ {
			s.tasks[level][j] = nil
		}

		s.tasks[level] = s.tasks[level][:count]
	} else if count > currentCount {
		extendBy := count - currentCount
		s.tasks[level] = append(s.tasks[level], make([]*Task, extendBy)...)
	}
}

func (s *Scheduler) sort() error {
	// sort the tasks into a layered topological order using the Coffman-Graham algorithm
	levels, err := s.graph.CoffmanGrahamSort(s.options.ConcurrentTasks)
	if err != nil {
		if err == dag.ErrCyclicGraph {
			return ErrCircularDependency
		}
		if err == dag.ErrDependencyOrder {
			return ErrOrderFailure
		}
		return err
	}

	// resize the top-level tasks slice to extend or shrink to the new number of levels
	s.resizeLevels(len(levels))

	for i, tasks := range levels {
		if s.tasks[i] == nil {
			s.tasks[i] = make([]*Task, len(tasks))
		} else {
			// resize the second-level tasks slice to extend or shrink to the new
			// number of tasks within the set
			s.resizeTasks(i, len(tasks))
		}

		for j, task := range tasks {
			s.tasks[i][j] = task.(*Task)
		}
	}

	return nil
}

// Run executes the scheduler's tasks.
func (s *Scheduler) Run(ctx context.Context) error {
	// check whether the scheduler's state has changed since the last time
	// it was ran, in which we case we need to sort the task graph
	if s.dirty {
		if err := s.sort(); err != nil {
			return err
		}

		s.dirty = false
	}

	// run each set of tasks, the count of the tasks within each set is guaranteed
	// to be equal or less than options.ConcurrentTasks
	for _, set := range s.tasks {
		if err := s.runner.Run(ctx, set); err != nil {
			return err
		}
	}

	return nil
}
