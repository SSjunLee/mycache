package singleflight

import "sync"

//call 代表正在进行中，或已经结束的请求
type call struct {
	wg  sync.WaitGroup
	val interface{} //这次调用的结果
	err error       //调用产生的错误
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call //不同 key 的请求(call)。
}

/**
针对相同的 key，无论 Do 被调用多少次，函数 `fn` 都只会被调用一次，等待 fn 调用结束了，返回返回值或错误。
**/
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = map[string]*call{}
	}

	if c, ok := g.m[key]; ok {
		//说明此时接口被调用，应该等待
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	return c.val, c.err
}
