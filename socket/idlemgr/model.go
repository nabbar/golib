/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package idlemgr

import (
	"context"
	"sync"
	"time"

	librun "github.com/nabbar/golib/runner"
	runtck "github.com/nabbar/golib/runner/ticker"
)

const (
	numShards = 32
	offset32  = 2166136261
	prime32   = 16777619
)

type shard struct {
	m sync.RWMutex      // mutex for this shard
	c map[string]Client // clients in this shard
}

type mgr struct {
	x context.Context  // context to stop running if trigger Done
	s [numShards]shard // shards for client storage
	r runtck.Ticker    // runner/Ticker instance
	i uint32           // idle timeout
}

// hashFNV1a implements the 32-bit FNV-1a hash algorithm inline to avoid allocations.
func hashFNV1a(s string) uint32 {
	var h uint32 = offset32
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= prime32
	}
	return h
}

func (o *mgr) getShard(ref string) *shard {
	return &o.s[hashFNV1a(ref)%numShards]
}

func (o *mgr) Start(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}
	return o.r.Start(ctx)
}

func (o *mgr) Stop(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}
	return o.r.Stop(ctx)
}

func (o *mgr) Restart(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}
	return o.r.Restart(ctx)
}

func (o *mgr) IsRunning() bool {
	if o == nil {
		return false
	}
	return o.r.IsRunning()
}

func (o *mgr) Uptime() time.Duration {
	if o == nil {
		return 0
	}
	return o.r.Uptime()
}

func (o *mgr) Register(c Client) error {
	if o == nil {
		return ErrInvalidInstance
	}

	if c == nil {
		return ErrInvalidClient
	}

	s := o.getShard(c.Ref())
	s.m.Lock()
	defer s.m.Unlock()

	if s.c == nil {
		s.c = make(map[string]Client)
	}

	s.c[c.Ref()] = c
	return nil
}

func (o *mgr) Unregister(c Client) error {
	if o == nil {
		return ErrInvalidInstance
	}

	if c == nil {
		return ErrInvalidClient
	}

	o.cleanRef(c.Ref())
	return nil
}

func (o *mgr) Close() error {
	for i := 0; i < numShards; i++ {
		s := &o.s[i]
		s.m.Lock()
		for k, c := range s.c {
			if c != nil {
				_ = c.Close()
			}
			delete(s.c, k)
		}
		s.m.Unlock()
	}

	return o.Stop(context.Background())
}

func (o *mgr) cleanRef(ref string) {
	s := o.getShard(ref)
	s.m.Lock()
	defer s.m.Unlock()

	if c, ok := s.c[ref]; ok {
		if c != nil {
			_ = c.Close()
		}
		delete(s.c, ref)
	}
}

func (o *mgr) run(ctx context.Context, _ librun.TickUpdate) error {
	if o == nil {
		return ErrInvalidInstance
	}

	// We iterate over all shards. Since this happens once per second,
	// the overhead of locking each shard sequentially is minimal.
	for i := 0; i < numShards; i++ {
		s := &o.s[i]
		var del []string

		s.m.RLock()
		if len(s.c) == 0 {
			s.m.RUnlock()
			continue
		}

		for k, c := range s.c {
			if c == nil {
				del = append(del, k)
				continue
			}

			c.Inc()

			if c.Get() > o.i {
				del = append(del, k)
			}
		}
		s.m.RUnlock()

		if len(del) > 0 {
			go func(sh *shard, keys []string) {
				sh.m.Lock()
				defer sh.m.Unlock()
				for _, k := range keys {
					if c, ok := sh.c[k]; ok {
						if c != nil {
							_ = c.Close()
						}
						delete(sh.c, k)
					}
				}
			}(s, del)
		}
	}

	if o.x.Err() != nil {
		go func() {
			_ = o.Stop(ctx)
		}()
		return o.x.Err()
	}

	return nil
}
