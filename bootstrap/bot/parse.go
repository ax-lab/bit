package bot

import (
	"errors"
	"fmt"

	"axlab.dev/bit/input"
)

type Parser func(ctx ParseContext, nodes NodeList)

func Parse(nodes NodeList) (errs []error) {
	parserList := []Parser{
		ParseBrackets,
		ParseLines,
		ParsePrint,
	}
	queueNext := []NodeList{nodes}
	for _, parser := range parserList {
		queue := queueNext
		queueNext = nil

		for _, nodes := range queue {
			ctx := parseContext{}
			parser(&ctx, nodes)
			queueNext = append(queueNext, ctx.queued...)
			if len(ctx.output) > 0 {
				nodes.data.Override(ctx.output)
				queueNext = append(queueNext, nodes)
			}

			if len(ctx.errs) > 0 {
				errs = append(errs, ctx.errs...)
			}
		}

		if len(errs) > 0 {
			break
		}
	}

	return
}

type ParseContext interface {
	Queue(nodes NodeList)
	Push(nodes ...Node)
	Error(err error)
	ErrorAt(span input.Span, msg string, args ...any)
}

type parseContext struct {
	errs   []error
	output []Node
	queued []NodeList
}

func (ctx *parseContext) Queue(nodes NodeList) {
	ctx.queued = append(ctx.queued, nodes)
}

func (ctx *parseContext) Push(nodes ...Node) {
	ctx.output = append(ctx.output, nodes...)
}

func (ctx *parseContext) Error(err error) {
	ctx.errs = append(ctx.errs, err)
}

func (ctx *parseContext) ErrorAt(span input.Span, msg string, args ...any) {
	var err error
	if len(args) > 0 {
		err = fmt.Errorf(msg, args...)
	} else {
		err = errors.New(msg)
	}
	ctx.Error(span.ErrorAt(err))
}
