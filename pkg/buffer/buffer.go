/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package buffer

import (
	"github.com/alipay/sofa-mosn/pkg/types"
	"sync"
	"errors"
	"context"
	"runtime"
	"sync/atomic"
)

const maxPoolSize = 1

// Register the bufferpool's name
const (
	Protocol     = iota
	SofaProtocol
	Stream
	SofaStream
	Proxy
	Bytes
	End
)

//
//var nullBufferCtx [End]interface{}
//
var (
	index    uint32
	poolSize = runtime.NumCPU()
	//bufferPoolContainers    [maxPoolSize]bufferPoolContainer
	//bufferPoolCtxContainers [maxPoolSize]bufferPoolCtxContainer
)
//
//type bufferPoolContainer struct {
//	pool [End]bufferPool
//}
//
//// bufferPool is buffer pool
//type bufferPool struct {
//	ctx types.BufferPoolCtx
//	sync.Pool
//}
//
//type bufferPoolCtxContainer struct {
//	sync.Pool
//}
//
//// Take returns a buffer from buffer pool
//func (p *bufferPool) take() (value interface{}) {
//	value = p.Get()
//	if value == nil {
//		value = p.ctx.New()
//	}
//	return
//}
//
//// Give returns a buffer to buffer pool
//func (p *bufferPool) give(value interface{}) {
//	p.ctx.Reset(value)
//	p.Put(value)
//}
//
//// PoolCtx is buffer pool's context
//type PoolCtx struct {
//	*bufferPoolContainer
//	value    [End]interface{}
//	transmit [End]interface{}
//}
//
//// NewBufferPoolContext returns a context with PoolCtx
//func NewBufferPoolContext(ctx context.Context, copy bool) context.Context {
//	if copy {
//		bufferCtx := PoolContext(ctx)
//		return context.WithValue(ctx, types.ContextKeyBufferPoolCtx, bufferCtxCopy(bufferCtx))
//	}
//
//	return context.WithValue(ctx, types.ContextKeyBufferPoolCtx, newBufferPoolCtx())
//}
//
//// TransmitBufferPoolContext copy a context
//func TransmitBufferPoolContext(dst context.Context, src context.Context) {
//	sctx := PoolContext(src)
//	if sctx.value == nullBufferCtx {
//		return
//	}
//	dctx := PoolContext(dst)
//	dctx.transmit = sctx.value
//	sctx.value = nullBufferCtx
//}
//
func bufferPoolIndex() int {
	i := atomic.AddUint32(&index, 1)
	i = i % uint32(maxPoolSize) % uint32(poolSize)
	return int(i)
}

//
//// newBufferPoolCtx returns PoolCtx
//func newBufferPoolCtx() (ctx *PoolCtx) {
//	i := bufferPoolIndex()
//	value := bufferPoolCtxContainers[i].Get()
//	if value == nil {
//		ctx = &PoolCtx{
//			bufferPoolContainer: &bufferPoolContainers[i],
//		}
//	} else {
//		ctx = value.(*PoolCtx)
//	}
//	return
//}
//
//func initBufferPoolCtx(poolCtx types.BufferPoolCtx) {
//	for i := 0; i < maxPoolSize; i++ {
//		pool := &bufferPoolContainers[i].pool[poolCtx.Name()]
//		pool.ctx = poolCtx
//	}
//}
//
//// GetPool returns buffer pool
//func (ctx *PoolCtx) getPool(poolCtx types.BufferPoolCtx) *bufferPool {
//	pool := &ctx.pool[poolCtx.Name()]
//	if pool.ctx == nil {
//		initBufferPoolCtx(poolCtx)
//	}
//	return pool
//}
//
//// Find returns buffer from PoolCtx
//func (ctx *PoolCtx) Find(poolCtx types.BufferPoolCtx, i interface{}) interface{} {
//	if ctx.value[poolCtx.Name()] != nil {
//		return ctx.value[poolCtx.Name()]
//	}
//	return ctx.Take(poolCtx)
//}
//
//// Take returns buffer from buffer pools
//func (ctx *PoolCtx) Take(poolCtx types.BufferPoolCtx) (value interface{}) {
//	pool := ctx.getPool(poolCtx)
//	value = pool.take()
//	ctx.value[poolCtx.Name()] = value
//	return
//}
//
//// Give returns buffer to buffer pools
//func (ctx *PoolCtx) Give() {
//	for i := 0; i < len(ctx.value); i++ {
//		value := ctx.value[i]
//		if value != nil {
//			ctx.pool[i].give(value)
//		}
//		value = ctx.transmit[i]
//		if value != nil {
//			ctx.pool[i].give(value)
//		}
//	}
//	ctx.transmit = nullBufferCtx
//	ctx.value = nullBufferCtx
//
//	i := bufferPoolIndex()
//
//	// Give PoolCtx to Pool
//	bufferPoolCtxContainers[i].Put(ctx)
//}
//
//func bufferCtxCopy(ctx *PoolCtx) *PoolCtx {
//	newctx := newBufferPoolCtx()
//	if ctx != nil {
//		newctx.value = ctx.value
//		ctx.value = nullBufferCtx
//	}
//	return newctx
//}
//
//// PoolContext returns PoolCtx by context
//func PoolContext(context context.Context) *PoolCtx {
//	if context != nil && context.Value(types.ContextKeyBufferPoolCtx) != nil {
//		return context.Value(types.ContextKeyBufferPoolCtx).(*PoolCtx)
//	}
//	return newBufferPoolCtx()
//}

const (
	defaultCapacity = 8
	// TODO calibrate
)

var (
	pool = sync.Pool{
		New: func() interface{} {
			return &bufferContext{make([]Reusable, 0, defaultCapacity)}
		},
	}

	ErrNoBufferCtx = errors.New("no buffer context found")
)

type Reusable interface {
	Free()
}

type bufferContext struct {
	// record reusable resources, mostly alloc from sync.Pool
	freeList []Reusable
}

func (c *bufferContext) append(reusable ... Reusable) {
	c.freeList = append(c.freeList, reusable...)
}

func (c *bufferContext) release() {
	// free resources
	for i, _ := range c.freeList {
		c.freeList[i].Free()
	}

	// reset free list
	c.freeList = c.freeList[:0]

	// give back to pool
	pool.Put(c)
}

func getBufferCtx(ctx context.Context) *bufferContext {
	if ctx != nil && ctx.Value(types.ContextKeyBufferCtx) != nil {
		return ctx.Value(types.ContextKeyBufferCtx).(*bufferContext)
	}
	return nil
}

func NewBufferCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, types.ContextKeyBufferCtx, pool.Get().(*bufferContext))
}

func GiveBufferCtx(ctx context.Context) {
	if ctx != nil && ctx.Value(types.ContextKeyBufferCtx) != nil {
		bufferCtx := ctx.Value(types.ContextKeyBufferCtx).(*bufferContext)

		// for debug
		//logger := log.ByContext(ctx)
		//connID := ctx.Value(types.ContextKeyConnectionID)
		//streamID := ctx.Value(types.ContextKeyStreamID)
		//logger.Debugf("conn %d, stream %d, release %d res", connID, streamID, len(bufferCtx.freeList))

		bufferCtx.release()
	}
}

func AppendReusable(ctx context.Context, reusable ... Reusable) error {
	bufferCtx := getBufferCtx(ctx)
	if bufferCtx == nil {
		return ErrNoBufferCtx
	}

	// for debug
	//logger := log.ByContext(ctx)
	//connID := ctx.Value(types.ContextKeyConnectionID)
	//streamID := ctx.Value(types.ContextKeyStreamID)
	//for i, _ := range reusable {
	//	logger.Debugf("conn %d, stream %d, append free list %+v", connID, streamID, reusable[i])
	//}

	bufferCtx.append(reusable...)
	return nil
}

func Move(src, dst context.Context) {
	srcBufferCtx := getBufferCtx(src)
	dstBufferCtx := getBufferCtx(dst)
	if srcBufferCtx == nil || dstBufferCtx == nil {
		return
	}

	// copy src to dst
	dstBufferCtx.freeList = append(dstBufferCtx.freeList, srcBufferCtx.freeList...)
	// for debug
	//logger := log.ByContext(dst)
	//connID := dst.Value(types.ContextKeyConnectionID)
	//streamID := dst.Value(types.ContextKeyStreamID)
	//for i, _ := range srcBufferCtx.freeList {
	//	logger.Debugf("conn %d, stream %d, move free list from us to ds %+v", connID, streamID, srcBufferCtx.freeList[i])
	//}

	// reset src free list
	srcBufferCtx.freeList = srcBufferCtx.freeList[:0]

	// give back to pool
	pool.Put(srcBufferCtx)
}
