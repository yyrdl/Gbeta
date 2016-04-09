package gbeta

import (
	"strings"
)

type Router struct {
	root           *_Node
	middlewares    []*_MiddlewareNode //store the middlewares
	current_max_id int                // it means how many handler node are added to the router
}

// set a middleware  in the "path"
func (r *Router) Use(path string, plugin Middlewares) {
	mw := new(_MiddlewareNode)
	mw.exsited_max_id = r.current_max_id - 1
	mw.path = strings.Join(formatPath(path), "/")
	if mw.path == "" {
		mw.path = "/"
	}
	mw.middleware = plugin
	r.middlewares = append(r.middlewares, mw)
}

func (r *Router) UseSubRouter(path string, subRouter *Router) {
	mergeRouterTo(subRouter, r, path)
}

func (r *Router) Post(path string, handler ReqHandler) {
	addHandler(r, path, "POST", handler)
}
func (r *Router) Put(path string, handler ReqHandler) {
	addHandler(r, path, "PUT", handler)
}
func (r *Router) Get(path string, handler ReqHandler) {
	addHandler(r, path, "GET", handler)
}
func (r *Router) Delete(path string, handler ReqHandler) {
	addHandler(r, path, "DELETE", handler)
}

func (r *Router) Options(path string, handler ReqHandler) {
	addHandler(r, path, "OPTIONS", handler)
}

func (r *Router) Patch(path string, handler ReqHandler) {
	addHandler(r, path, "PATCH", handler)
}

func addHandler(r *Router, path string, method string, handler ReqHandler) {
	node := new(_Node)         //create a new _Node
	node.id = r.current_max_id // set the id
	node.existed_more_op = false
	node.handler = handler
	node.method = method
	node.next = nil
	for i := 0; i < 12; i++ { //显示初始化，放心一点(~_~),最怕内存访问越界了
		node.children[i] = nil
	}
	paths := formatPath(path)
	path = strings.Join(paths, "/")
	if path == "" {
		path = "/"
	}
	node.middleware_to_excute = caculateMiddlewareToExcute(r, path, node.id)
	r.current_max_id += 1
	addNodeToRouter(r, paths, node)
}

//比较昂贵的操作，只在建路由树的时候执行
// caculate middleware to excute when a handler will been excuted

func caculateMiddlewareToExcute(r *Router, path string, node_id int) []int {
	var (
		result []int
		length int = 0
	)
	for i := 0; i < len(r.middlewares); i++ {
		if r.middlewares[i].exsited_max_id < node_id { //表明中间件先添加
			if r.middlewares[i].path == "/" {
				result = append(result, i)
			} else {
				length = len(r.middlewares[i].path)
				if length <= len(path) {
					if path[:length] == r.middlewares[i].path {
						if length == len(path) {
							result = append(result, i)
						} else {
							if path[length] == 47 || path[length] == 63 { //'/'或"?"
								result = append(result, i)
							}
						}
					}
				}
			}
		}
	}
	return result
}

//return a new Router(*Router)
func NewRouter() *Router {
	router := new(Router)
	router.root = new(_Node)
	router.current_max_id = 0
	return router
}
