// by yyrdl ,MIT License
package gbeta

import (
	"strings"
)

//定义_Path，而不直接是string对比的原因在于为以后再路径中添加正则表达式留接口

type _Path struct {
	path    string
	more_op bool
}

//maybe will rewrite the path code to support path param or regexp
func newPath(path string) *_Path {
	var p = new(_Path)
	p.path = path
	p.more_op = needMoreOp(path)
	return p
}

//more_op表示该路径不能只是简单的比较字符串，可能需要正则
func needMoreOp(path string) bool {
	if len(path) > 0 {
		if path[:1] == ":" { //路径寻找的时候用到路由上路径的长度，若路由上出现参数则会出错
			return true
		}
	}
	return false
}

//若日后想要添加正则表达式功能直接修改这个path的相应函数即可

func (p *_Path) Match(subPath string) (is_match bool, key string, value string, left_path string) {
	if p.path[:1] == ":" {
		var i int = 0
		for ; i < len(subPath); i++ {
			if subPath[i] == 47 { //"/"
				return true, p.path[1:], subPath[:i], subPath[i+1:]
			}
			if subPath[i] == 63 { // "?"
				return true, p.path[1:], subPath[:i], ""
			}
		}
		return true, p.path[1:], subPath, ""
	}
	return false, "", "", subPath
}

//在pathNode.go的addNode函数里用到，用来比较俩个子路径是否一致
func (p *_Path) IsSimillar(path string) bool {
	return p.path == path
}

//在生成路由树的时候用到，格式化路径，避免"path/p" "/sa///asa/"这样的路径出现
//这是最耗时的操作，在路由请求时不再会用到，路由请求时默认req.URL.Path是标准的
//对"ds/ds","/ds/ds","//ds//ds"都将输出[]string{"","ds","ds"}

func formatPath(url string) []string {
	var (
		start int = 0
		num   int = 1
	)
	url = strings.Join(strings.Split(url, " "), "")
	temp := make([]string, 10)
	temp[0] = ""
	for i := 0; i < len(url); i++ {
		if url[i:i+1] == "/" {
			if start != i {
				if num < 10 {
					temp[num] = url[start:i]
				} else {
					temp = append(temp, url[start:i])
				}
				num += 1
			}
			start = i + 1
			continue
		}
		if i == len(url)-1 {
			if start != i {
				if num < 10 {
					temp[num] = url[start:]
				} else {
					temp = append(temp, url[start:])
				}
				num += 1
			}
			break
		}
		if url[i:i+1] == "?" {
			if start != i {
				if num < 10 {
					temp[num] = url[start:i]
				} else {
					temp = append(temp, url[start:i])
				}
				num += 1
			}
			break
		}
	}
	return temp[:num]
}
