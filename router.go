package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

//This is a simple restful router for a go http server

//Explicity error type for the router
type RouterError struct {
	err error
}

func (r *RouterError) Error() string {
	return r.err.Error()
}

//Easy constructor for RouterError
func NewRouterError(message string, params ...interface{}) (r *RouterError) {
	r = &RouterError{err: fmt.Errorf(message, params...)}
	return
}

//Struct to pass parsed rest params
type RestParams struct {
	pathParams map[string]string
	mu         sync.RWMutex
}

func NewRestParams() (rp *RestParams) {
	rp = &RestParams{pathParams: make(map[string]string)}
	return
}

func (rp *RestParams) Get(param string) string {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	return rp.pathParams[param]
}

func (rp *RestParams) set(param string, value string) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	rp.pathParams[param] = value
}

//this is a route entry that the server will match against on incoming http requests
type routeEntry struct {
	path       *regexp.Regexp
	method     string
	pathParams []string
	handler    func(http.ResponseWriter, *http.Request, *RestParams)
}

//the rest router is a map of request method to accepted paths
type RestRouter struct {
	mu     sync.RWMutex
	routes map[string][]routeEntry //method to route entry
}

//parse the requested path, variables and pass to the handler
func (rest *RestRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rest.mu.RLock()
	defer rest.mu.RUnlock()

	var routes []routeEntry
	var ok bool

	if routes, ok = rest.routes[r.Method]; !ok {
		rw.WriteHeader(404)
		return
	}

	for _, route := range routes {
		if route.path.MatchString(r.URL.Path) {
			matches := route.path.FindStringSubmatch(r.URL.Path)
			if len(matches) != len(route.pathParams)+1 {
				rw.WriteHeader(400)
				return
			}

			restParams := NewRestParams()

			for i, paramName := range route.pathParams {
				restParams.set(paramName, matches[1+i])
			}

			route.handler(rw, r, restParams)
			return
		}
	}

	rw.WriteHeader(404)
}

//turn a url pattern for rest into a regex and path params
func makePath(urlPattern string) (regex *regexp.Regexp, pathParams []string, err error) {
	pathParams = make([]string, 0)
	//patternPeices = make([]string)
	parts := strings.Split(urlPattern, "/")
	for i, p := range parts {
		l := len(p)
		if l > 0 && p[0] == ':' {
			if l < 2 {
				//unnamed param
				err = NewRouterError("Cannot have unnamed path params in route: ", urlPattern)
				return
			}
			pathParams = append(pathParams, p[1:l])
			parts[i] = "([A-Za-z0-9]+)" //just support simple strings not url encoded or anything for now
		}
	}

	regexString := "^" + strings.Join(parts, "/") + "$"
	regex, err = regexp.Compile(regexString)

	return
}

//exported function to add a route
func (rest *RestRouter) AddRoute(method string, urlPattern string, handler func(http.ResponseWriter, *http.Request, *RestParams)) (err error) {
	rest.mu.Lock()
	defer rest.mu.Unlock()

	if method == "" || urlPattern == "" {
		err = NewRouterError("Cannot have an empty pattern or method")
		return
	}

	if handler == nil {
		err = NewRouterError("Cannot have a nil handler")
		return
	}

	re := routeEntry{}
	re.handler = handler
	re.path, re.pathParams, err = makePath(urlPattern)

	if err != nil {
		return err
	}

	if _, ok := rest.routes[method]; !ok {
		rest.routes[method] = make([]routeEntry, 0)
	}
	rest.routes[method] = append(rest.routes[method], re)

	return
}

//easy constructor
func NewRestRouter() (rest *RestRouter) {
	rest = &RestRouter{routes: make(map[string][]routeEntry)}
	return
}
