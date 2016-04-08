//by yyrdl ,MIT License . welcome to use and welcome to star it :) Issues are welcome too!

package gbeta

import (
	"errors"
	"strings"
)

//新的树形结构，在生成树时做更多工作，可以在路由寻址时更快。主要改动有俩点：
// 1. 在节点上加索引 ，%操作平均每个7ns
// 2. 在生成路由树阶段就确定下该节点应该要依次执行的中间件

type _Node struct {
	id       int        //only node with handle will be setted
	children [12]*_Node //0-9放一般的节点，10放需要特殊处理的节点，若一个节点的子节点包含需要特殊
	//的节点处理的，则优先匹配，11存放len(pointer.path.path)==0的节点
	next                 *_Node     //下一个节点（兄弟节点）,link to sibling
	existed_more_op      bool       //是否存在需优先处理的节点,如path param或者正则式
	handler              ReqHandler //具体请求的执行句柄
	method               string     //post put get ....
	path                 *_Path     // subpath 一个_Path
	middleware_to_excute []int      //需要执行的Middleware ,middleware to excute when the handler will be excuted
}

func createBlankNode(subpath string) *_Node {
	node := new(_Node)
	node.path = newPath(subpath)
	node.next = nil
	node.handler = nil
	node.method = ""
	node.existed_more_op = false
	for i := 0; i < 12; i++ {
		node.children[i] = nil
	}
	return node
}

/*******************************Add _Node*********************************************/
//在创建路由时我们可以耗一点时间，尽量容错，尽量多做点工作，之后路由寻址的时候就会快很多
//在合并俩个router的时候直接把子router挂在父节点上即可，同时应该更新middleware和子router中handler
//节点的id及其需要执行的middleware
// add a new handler to the router
func addNodeToRouter(r *Router, paths []string, node *_Node) {
	//作为Router其直接孩子的路径应是以"/"开头,formatPath（）之后,paths[0]应为""
	if paths[0] != "" {
		panic(errors.New("Failed to add node to router ,the path should begin with'/'"))
	} else {
		err, exist_path, _ := recursionAddNode(r.root, 0, paths, node)
		if err {
			panic(errors.New("Failed to add the same method :'" + node.method + "' to the same path:" + exist_path))
		}
	}
}

//递归添加_Node
// recursion add node
//if the node with same path and methos is existing,the 'wrong' will be true
func recursionAddNode(father *_Node, i int, paths []string, node *_Node) (wrong bool, existed_path string, existed_node *_Node) {
	if len(paths) == i {
		return false, "", nil
	}
	var (
		pointer *_Node
		forward *_Node
		index   byte = 0
	)

	if needMoreOp(paths[i]) { //whether the subpath is a special path ,like ":username" or regexp expression
		father.existed_more_op = true
		index = 10
	} else {
		if len(paths[i]) == 0 {
			index = 11
		} else {
			index = paths[i][0] % 10 //在我的机器上(i5),这个操作平均 7ns
		}
	}

	pointer = father.children[index]

	if pointer == nil { //index索引下为空
		if i == len(paths)-1 { //当前是最后一个路径
			node.path = newPath(paths[i])
			node.next = nil
			father.children[index] = node
			return false, "", nil
		} else {
			b_node := createBlankNode(paths[i])
			father.children[index] = b_node
			return recursionAddNode(b_node, i+1, paths, node)
		}
	}

	for {
		if pointer == nil {
			break
		}
		if pointer.path.IsSimillar(paths[i]) { //存在一个相似索引
			if i == len(paths)-1 { //如果当前是最后一个路径
				if node.handler == nil { //方便合并router,support for  router merging
					//node.handler若是nil的话，只能是在合并router时传入的一个空node
					//此时存在一个相同的路径，返回已经存在的末节点的指针
					return true, "", pointer
				} else { //否则为正常的添加handler节点
					if pointer.handler != nil && pointer.method == node.method {
						//节点冲突
						return true, "/" + paths[i], nil
					} else {
						if pointer.handler == nil { //为空节点，复制node到空节点
							pointer.handler = node.handler
							pointer.id = node.id
							pointer.middleware_to_excute = node.middleware_to_excute
							return false, "", nil //成功复制到空节点
						} else { //允许对同一路径添加不同method的handler
							forward = pointer
							pointer = pointer.next
							continue
						}
					}
				}
			} else { //否则直接转到下层,不管pointer是个什么节点
				return recursionAddNode(pointer, i+1, paths, node)
			}
		} else {
			forward = pointer
			pointer = pointer.next
		}
	}
	//上面没返回，则挂载到后面
	if i == len(paths)-1 { //此时无论是空节点还是handler节点，操作都是一样的
		node.path = newPath(paths[i])
		node.next = nil
		forward.next = node
		return false, "", nil
	} else {
		b_node := createBlankNode(paths[i])
		b_node.next = nil
		forward.next = b_node
		return recursionAddNode(b_node, i+1, paths, node)
	}
}

/***************************Merge Router*******************************************/
//包含链接俩个router，更新node的ID,更新middlerware的exsited_max_id
//更新node.middleware_to_excute的每个值,只是在原来的基础上加一个数
//更新时需注意合并之后是否存在冲突的路径
// merge a  router 'from' to another router 'to'
func mergeRouterTo(from *Router, to *Router, path string) {

	var (
		father    *_Node
		pointer   *_Node
		id_offset int = to.current_max_id
	)
	//第一步：首先在to 上找到path(存在则不添加，不存在则添加)
	paths := formatPath(path)
	if paths[0] != "" {
		panic(errors.New("Failed to merge router ,the path should begin with'/'"))
	}

	b_node := createBlankNode("")
	b_node.path = nil
	//寻找挂载点
	// find the node where to link the subrouter
	exist, _, pointer := recursionAddNode(to.root, 0, paths, b_node)
	if exist { //若该路径下存在节点，则pointer是该路径下出现的第一个节点，虽然允许
		//对同一个路径添加不同method的handler，也允许在handler后面再添加子节点，只要保证
		//子节点添加到第一个出现的同path的handler上,之后的寻址算法就默认从第一个去找，避免丢失
		father = pointer
	} else { //不存在，则已经添加进去，即为b_node
		father = b_node
	}

	//update the information of middlewares on the subrouter
	for i := 0; i < len(from.middlewares); i++ {
		from.middlewares[i].path = strings.Join(formatPath(path+"/"+from.middlewares[i].path), "/")
		from.middlewares[i].exsited_max_id += id_offset
	}

	// add the middlewares of subrouter to father router
	to.middlewares = append(to.middlewares, from.middlewares...)
	//get the father_path
	father_path := strings.Join(paths, "/")
	if father_path == "" {
		father_path = "/"
	}
	//update the node id and middleware to excute
	updateNodeChildren(to, from.root, id_offset, father_path)
	// then start merging
	if from.root.children[11] != nil { //子路由的节点一定是在children[11]
		pointer = from.root.children[11]
		//merge the root of subrouter to father ,pay attention to same method on the same path!!!
		mergeRoot(father, pointer, path) //
		pointer = from.root.children[11]

		for i := 0; i < 12; i++ {
			if father.children[i] == nil { //father上该处为nil,简单复制过来就OK
				father.children[i] = pointer.children[i]
			} else {
				mergeSubTree(pointer.children[i], father.children[i])
			}
		}
	}
}

//递归同步走path,from是一条链，to也是一条链
func mergeSubTree(from *_Node, to *_Node) {
	var (
		pointer *_Node = from
		target  *_Node
		forward *_Node
		done    bool = false
	)

	for { //第一个for循环走from
		if pointer == nil {
			break
		}
		target = to
		for { //第二个循环走to
			if target == nil {
				break
			}
			//如果发现相同路径，
			if target.path.IsSimillar(pointer.path.path) {
				if pointer.handler == nil { //若此时pointer是一个空白节点，直接合并孩子
					mergeChildren(pointer, target)
					done = true
					break
				} else { //是一个handler节点，需要看target是什么节点
					if target.handler == nil { //复制pointer到target，合并孩子
						target.handler = pointer.handler
						target.method = pointer.method
						target.id = pointer.id
						target.middleware_to_excute = pointer.middleware_to_excute
						if pointer.existed_more_op {
							target.existed_more_op = true
						}
						mergeChildren(pointer, target)
						done = true
						break
					} else {
						if target.method == pointer.method {
							panic("Failed to merge subtree,same method on the same path!")
						} else {
							forward = target
							target = target.next
						}
					}
				}
			} else {
				forward = target
				target = target.next
			}
		}

		if !done {
			forward.next = pointer
		}
		pointer = pointer.next

		if !done && forward != nil && forward.next != nil && forward.next.next != nil {
			forward.next.next = nil
		}
		done = false
	}
	return
}

func mergeChildren(from *_Node, to *_Node) {
	if from.existed_more_op {
		to.existed_more_op = true
	}
	for i := 0; i < 12; i++ {
		if to.children[i] == nil {
			to.children[i] = from.children[i]
		} else {
			if from.children[i] == nil {
				continue
			} else {
				mergeSubTree(from.children[i], to.children[i])
			}
		}
	}
	return
}

//合并子router的root到father节点，并且注意同路径多handler的情况
// merge subrouter's root to his father node on another router
// pay attention to mutiple handler on the same path with different method
func mergeRoot(start *_Node, next *_Node, path string) {
	var (
		pointer *_Node
		forward *_Node
	)

	if next.handler != nil { //在路径"/"上就已经绑定了handler,可能还是多个
		if start.handler == nil {
			start.handler = next.handler
			start.id = next.id
			start.method = next.method
			if next.existed_more_op {
				start.existed_more_op = next.existed_more_op
			}
			start.middleware_to_excute = next.middleware_to_excute
			next = next.next
		}

		for {
			pointer = start
			if next == nil {
				break
			}
			for {
				if pointer == nil {
					break
				}
				if pointer.method == next.method {
					panic(errors.New("Faild to add the same method'" + next.method + "' to the same path:" + path))
				} else {
					forward = pointer
					pointer = pointer.next
				}
			}
			forward.next = next

			next = next.next

			if forward.next != nil {
				forward.next.next = nil
			}
		}

	} //如果next.handler是nil的话，那么就只是一个单独的空节点
}

//递归更新子router的handler node
//update the id and the middlewares_to_excute on the subrouter
func updateNodeChildren(to *Router, father *_Node, id_offset int, father_path string) {
	for i := 0; i < 12; i++ {
		if father.children[i] != nil {
			updateNodeList(to, father.children[i], id_offset, father_path)
		}
	}
	return
}

func updateNodeList(to *Router, start *_Node, id_offset int, father_path string) {
	var pointer *_Node = start
	for {
		if pointer == nil {
			break
		}
		if pointer.handler != nil {
			pointer.id += id_offset
			//重新计算需要执行的中间件
			pointer.middleware_to_excute = caculateMiddlewareToExcute(to, father_path+"/"+pointer.path.path, pointer.id)
		}
		updateNodeChildren(to, pointer, id_offset, father_path+"/"+pointer.path.path)
		pointer = pointer.next
	}
	return
}

/****************************Find _Node******************************************/
//根据req.URL.Path和method寻找指定的node，在树的结构确定之后，这个函数将是性能的关键
//len() 0.88ns/op if 1.44ns/op == 1.73ns/op
//单独测试树形结构访问四层只需 33ns/op 为什么到这里需要244ns/op????

// find the node on the path with particular method
// the func will return the node's pointer ,and if there are any path parameters,the
//func will return them as an string ,such as "key1,value1;key2,value2;"
func findNode(start *_Node, path string, method string) (*_Node, string) {
	var (
		kvs       string = "" //存放path parameter
		match     bool   = false
		key       string = ""
		value     string = ""
		length    int    = 0
		path_left string = ""
	)

	for {
		if start == nil { //< 5ns/op
			break
		}
		length = len(start.path.path) //1ns/op
		match = false
		if start.path.more_op { //whether the subpath is an special path ,and need more operation
			match, key, value, path_left = start.path.Match(path) //then do more operation by it self
			if key != "" {
				kvs = kvs + key + "," + value + ";"
			}

		} else {
			if len(path) >= length { //then just see if they are equal
				match = (start.path.path == path[:length]) && (len(path) == length || (path[length] == 47 || path[length] == 63))
				if match && len(path) != length {
					path_left = path[length+1:]
				} else {
					path_left = ""
				}
			} else {
				start = start.next
				continue
			}
		}
		if match {
			if len(path_left) > 0 && path[length] != 63 { //"?"决定是否继续往下找
				// judge if we should search deeper
				if start.existed_more_op {
					start = start.children[10]
				} else {
					start = start.children[path_left[0]%10]
				}
				path = path_left
				continue
			} else {
				if start.method == method { //若method匹配
					return start, kvs
				}
			}
		}
		start = start.next
	}
	return nil, ""
}

//寻找某一个路径下允许使用的method 结果形如"POST,PUT"
//find the allowed http methods on the given path
func findAllowedMethod(start *_Node, path string) string {
	var (
		match     bool   = false
		length    int    = 0
		path_left string = ""
		methods   string = ""
	)

	for {
		if start == nil { //< 5ns/op
			break
		}
		length = len(start.path.path) //1ns/op
		match = false
		if start.path.more_op {
			match, _, _, path_left = start.path.Match(path)
		} else {
			if len(path) >= length {
				match = (start.path.path == path[:length]) && (len(path) == length || (path[length] == 47 || path[length] == 63))
				if match && len(path) != length {
					path_left = path[length+1:]
				} else {
					path_left = ""
				}
			} else {
				start = start.next
				continue
			}
		}
		if match {
			if len(path_left) > 0 && path[length] != 63 { //"?"决定是否继续往下找
				if start.existed_more_op {
					start = start.children[10]
				} else {
					start = start.children[path_left[0]%10]
				}
				path = path_left
				continue
			} else {
				if start.handler != nil {
					methods = methods + start.method + ","
				}
			}
		}
		start = start.next
	}
	if methods == "" {
		return methods
	} else {
		return methods[:len(methods)-1]
	}
}
