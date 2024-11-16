package gee

import (
	"log"
	"net/http"
	"strings"
)

// HandlerFunc HandlerFunc定义了一个函数
type HandlerFunc func(c *Context)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *RouterGroup
	engine      *Engine
}

// Engine Engine类定义了一个路由映射表
type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup
}

func New() *Engine {
	engine := Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: &engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return &engine
}
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}
func (group *RouterGroup) Use(middleware ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middleware...)
}
func (group *RouterGroup) addRoute(method string, comb string, handler HandlerFunc) {
	pattern := group.prefix + comb
	log.Println("Route %4s-%s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET 定义用get方法添加请求
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST 定义post方法添加请求
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// Run 开始一个http server服务器
func (engine *Engine) Run(addr string) (err error) {
	//这里实际上有一个engine到Handler类型的自动转换，因为engine实现了Handler接口中的方法，所有可以转换为Handler类型
	return http.ListenAndServe(addr, engine)
}

// 实现Handler接口的ServeHTTP方法
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	c := newContext(w, req)
	c.handlers = append(c.handlers, middlewares...)
	engine.router.handle(c)
}
