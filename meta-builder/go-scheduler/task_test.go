package go_scheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskDependencies(t *testing.T) {
	scheduler := New()

	a := newTask(scheduler, testSchedulerTask)
	assert.Empty(t, a.Dependencies(), "a.Dependencies() should equal nil")
	assert.Zero(t, a.DependencyCount(), "a.DependencyCount() should equal 0")

	b := newTask(scheduler, testSchedulerTask)
	a.DependsOn(b)

	assert.Equal(t, []*Task{b}, a.Dependencies(), "a.Dependencies() should equal [b]")
	assert.Equal(t, 1, a.DependencyCount(), "a.DependencyCount() should equal 1")

	a.RemoveDependency(b)

	assert.Empty(t, a.Dependencies(), "a.Dependencies() should equal nil")
	assert.Zero(t, a.DependencyCount(), "a.DependencyCount() should equal 0")
}
