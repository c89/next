// Copyright (c) 2015 Fred Zhou

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package next

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	Config *Config
	routes *Routes
	Logger *log.Logger
	Env    map[string]interface{}
	//save the listener so it can be closed
	l net.Listener
}

func NewServer() *Server {
	server := &Server{
		Config: NewConfig(),
		routes: NewRoutes(),
		Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
		Env:    map[string]interface{}{},
	}

	// Load default config if exists
	file := "config.json"
	if fileExists(file) {
		server.Config.Load(file)
	}

	return server
}

// ServeHTTP is the interface method for Go's http server package
func (s *Server) ServeHTTP(c http.ResponseWriter, req *http.Request) {
	s.Process(c, req)
}

// Process invokes the routing system for server s
func (s *Server) Process(c http.ResponseWriter, req *http.Request) {
	route := s.routeHandler(req, c)
	if route != nil {
		route.httpHandler.ServeHTTP(c, req)
	}
}

// Get adds a handler for the 'GET' http method for server s.
func (s *Server) Get(route string, handler interface{}) {
	s.routes.Add(route, "GET", handler)
}

// Post adds a handler for the 'POST' http method for server s.
func (s *Server) Post(route string, handler interface{}) {
	s.routes.Add(route, "POST", handler)
}

// Put adds a handler for the 'PUT' http method for server s.
func (s *Server) Put(route string, handler interface{}) {
	s.routes.Add(route, "PUT", handler)
}

// Delete adds a handler for the 'DELETE' http method for server s.
func (s *Server) Delete(route string, handler interface{}) {
	s.routes.Add(route, "DELETE", handler)
}

// Match adds a handler for an arbitrary http method for server s.
func (s *Server) Match(method string, route string, handler interface{}) {
	s.routes.Add(route, method, handler)
}

//Adds a custom handler. Only for webserver mode. Will have no effect when running as FCGI or SCGI.
func (s *Server) Handler(route string, method string, handler http.Handler) {
	s.routes.Add(route, method, handler)
}

// Run starts the web application and serves HTTP requests for s
func (s *Server) Run(addr string) {
	mux := http.NewServeMux()
	if s.Config.Bool("debug.profiler") {
		mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
		mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	}
	mux.Handle("/", s)

	s.Logger.Printf("next serving %s\n", addr)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
	s.l = l
	err = http.Serve(s.l, mux)
	s.l.Close()
}

// RunTLS starts the web application and serves HTTPS requests for s.
func (s *Server) RunTLS(addr string, config *tls.Config) error {
	mux := http.NewServeMux()
	mux.Handle("/", s)
	l, err := tls.Listen("tcp", addr, config)
	if err != nil {
		log.Fatal("Listen:", err)
		return err
	}

	s.l = l
	return http.Serve(s.l, mux)
}

// Close stops server s.
func (s *Server) Close() {
	if s.l != nil {
		s.l.Close()
	}
}

// safelyCall invokes `function` in recover block
func (s *Server) safelyCall(function reflect.Value, args []reflect.Value) (resp []reflect.Value, e interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if s.Config.Bool("panic") {
				// go back to panic
				panic(err)
			} else {
				e = err
				resp = nil
				s.Logger.Println("Handler crashed with error", err)
				for i := 1; ; i += 1 {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					s.Logger.Println(file, line)
				}
			}
		}
	}()
	return function.Call(args), nil
}

// requiresContext determines whether 'handlerType' contains
// an argument to 'web.Ctx' as its first argument
func requiresContext(handlerType reflect.Type) bool {
	//if the method doesn't take arguments, no
	if handlerType.NumIn() == 0 {
		return false
	}

	//if the first argument is not a pointer, no
	a0 := handlerType.In(0)
	if a0.Kind() != reflect.Ptr {
		return false
	}
	//if the first argument is a context, yes
	if a0.Elem() == reflect.TypeOf(Context{}) {
		return true
	}

	return false
}

func (s *Server) logRequest(ctx Context, sTime time.Time) {
	//log the request
	var logEntry bytes.Buffer
	req := ctx.Request
	requestPath := req.URL.Path

	duration := time.Now().Sub(sTime)
	var client string

	// We suppose RemoteAddr is of the form Ip:Port as specified in the Request
	// documentation at http://golang.org/pkg/net/http/#Request
	pos := strings.LastIndex(req.RemoteAddr, ":")
	if pos > 0 {
		client = req.RemoteAddr[0:pos]
	} else {
		client = req.RemoteAddr
	}

	fmt.Fprintf(&logEntry, "%s - \033[32;1m %s %s\033[0m - %v", client, req.Method, requestPath, duration)

	if len(ctx.Params) > 0 {
		fmt.Fprintf(&logEntry, " - \033[37;1mParams: %v\033[0m\n", ctx.Params)
	}

	ctx.Server.Logger.Print(logEntry.String())
}

// the main route handler in next
// Tries to handle the given request.
// Finds the route matching the request, and execute the callback associated
// with it.  In case of custom http handlers, this function returns an "unused"
// route. The caller is then responsible for calling the httpHandler associated
// with the returned route.
func (s *Server) routeHandler(req *http.Request, w http.ResponseWriter) (unused *Route) {
	requestPath := req.URL.Path
	ctx := Context{req, map[string]string{}, s, w}

	//set some default headers
	ctx.SetHeader("Server", "next", true)
	tm := time.Now().UTC()

	//ignore errors from ParseForm because it's usually harmless.
	req.ParseForm()
	if len(req.Form) > 0 {
		for k, v := range req.Form {
			ctx.Params[k] = v[0]
		}
	}
	if len(req.PostForm) > 0 {
		for k, v := range req.PostForm {
			ctx.Params[k] = v[0]
		}
	}

	// Data in body
	// ------------------
	requestbody, _ := ioutil.ReadAll(req.Body)
	req.Body.Close()
	bf := bytes.NewBuffer(requestbody)
	req.Body = ioutil.NopCloser(bf)

	if len(requestbody) > 0 {
		postArray := strings.Split(string(requestbody), "&")
		for _, v := range postArray {
			pair := strings.Split(v, "=")
			if len(pair) > 1 {
				ctx.Params[pair[0]], _ = url.QueryUnescape(pair[1])
			}
		}
	}
	// ------------------

	defer s.logRequest(ctx, tm)

	ctx.SetHeader("Date", webTime(tm), true)

	route := s.routes.Match(requestPath, req.Method)
	if route == nil {
		ctx.Abort(404, "Page not found")
		return
	}
	cr := route.cr

	//Set the default content-type
	ctx.SetHeader("Content-Type", "text/html; charset=utf-8", true)

	if route.httpHandler != nil {
		unused = route
		// We can not handle custom http handlers here, give back to the caller.
		return
	}

	var args []reflect.Value
	handlerType := route.handler.Type()
	if requiresContext(handlerType) {
		args = append(args, reflect.ValueOf(&ctx))
	}

	match := cr.FindStringSubmatch(requestPath)
	for _, arg := range match[1:] {
		args = append(args, reflect.ValueOf(arg))
	}

	ret, err := s.safelyCall(route.handler, args)
	if err != nil {
		//there was an error or panic while calling the handler
		ctx.Abort(500, "Server Error")
	}
	if len(ret) == 0 {
		return
	}

	sval := ret[0]

	var content []byte

	if sval.Kind() == reflect.String {
		content = []byte(sval.String())
	} else if sval.Kind() == reflect.Slice && sval.Type().Elem().Kind() == reflect.Uint8 {
		content = sval.Interface().([]byte)
	}
	ctx.SetHeader("Content-Length", strconv.Itoa(len(content)), true)
	_, err = ctx.ResponseWriter.Write(content)
	if err != nil {
		ctx.Server.Logger.Println("Error during write: ", err)
	}

	return
}

// SetLogger sets the logger for server s
func (s *Server) SetLogger(logger *log.Logger) {
	s.Logger = logger
}
