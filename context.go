//by yyrdl ,MIT License ,welcome to use and welcome to star it :)
package gbeta

import (
	"sync"
)

//context stores values of url parameters ,and maybe middleware will also set
// some key-value in it . Context can be passed to functions running in different
//goroutines.Contexts are safe for simultaneous use by multiple goroutines.
//Please just pass the pointer (*Context)， do not copy it

//context 用来存放url参数，中间件也可以在上面设置信息，比如json-parser可以将json的解析结果放在里面
//context是线程安全的，但一定要以指针的方式进行传递
type Context struct {
	mu    *sync.RWMutex
	store map[interface{}]interface{}
}

//设置一个键值对
//set a key-value
func (c *Context) Set(key, value interface{}) {
	c.mu.Lock()
	c.store[key] = value
	c.mu.Unlock()
}

//获取key对应的value
// get the value of corresponding key
func (c *Context) Get(key interface{}) interface{} {
	c.mu.RLock()
	v := c.store[key]
	c.mu.RUnlock()
	return v
}

//删除某一个键值对
//delete a key-value pair
func (c *Context) Delete(key interface{}) {
	c.mu.Lock()
	delete(c.store, key)
	c.mu.Unlock()
}

//按用户定义的check规则检查某一个值，如果返回true，则对应的key ，value添加进去
//if the checkFunc return true ,the value will be stored
func (c *Context) CheckAndSet(key, value interface{}, checkFunc CheckFunc) bool {
	is_successful := false
	c.mu.Lock()
	v := c.store[key]
	if checkFunc(v, value) {
		c.store[key] = value
		is_successful = true
	}
	c.mu.Unlock()
	return is_successful
}

//清空map ，以便放回sync.pool
//clear the key-value that are stored in the map
func (c *Context) Clear() {
	c.mu.Lock()
	for key, _ := range c.store {
		delete(c.store, key)
	}
	c.mu.Unlock()
}

//create a new Context
func NewContext() *Context {
	c := new(Context)
	c.mu = new(sync.RWMutex)
	c.store = make(map[interface{}]interface{}, 5)
	return c
}
