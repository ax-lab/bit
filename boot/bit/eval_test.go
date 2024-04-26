package bit_test

import (
	"testing"
	"time"

	"axlab.dev/bit/boot/bit"
	"github.com/stretchr/testify/require"
)

func TestEvalEmpty(t *testing.T) {
	test, ctx := newEvalTest(t)
	res := ctx.Eval()
	test.Empty(res.Errors())
}

func TestEvalSimple(t *testing.T) {
	test, ctx := newEvalTest(t)
	queue := ctx.Queue()

	numTasks := 10

	isDone := make([]bool, numTasks)

	for i := 0; i < numTasks; i++ {
		index := i
		task := queue.NewTask(func(task *bit.EvalTask) {
			<-time.After(time.Duration(10*(index+1)) * time.Millisecond)
			isDone[index] = true
		})
		task.Queue()
	}

	<-time.After(15 * time.Millisecond)
	test.False(isDone[0], "tasks started before context eval")

	res := ctx.Eval()
	test.Empty(res.Errors())

	for n, it := range isDone {
		test.True(it, "task #%d was not executed after context eval", n+1)
	}

}

func newEvalTest(t *testing.T) (*require.Assertions, *bit.Context) {
	test := require.New(t)
	comp := bit.Compiler{}
	ctx, err := comp.NewContext()
	test.NoError(err)
	return test, ctx
}
