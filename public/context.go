package public

import (
	"math"
	"github.com/ridewindx/mel"
)

const preStartIndex int8 = -1
const abortIndex int8 = math.MaxInt8 / 2

type Handler func(*Context)

type Context struct {
	*Event

	response interface{}

	index    int8
	handlers []Handler

	*mel.Context
}

func (c *Context) Next() {
	c.index++
	s := int8(len(c.handlers))
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Abort() {
	c.index = abortIndex
}

func (c *Context) WriteResponse(rep interface{}) {
	c.response = rep
}
