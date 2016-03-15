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
	"crypto/tls"
	"net/http"
)

const VERSION = "0.0.1"

// Process invokes the main server's routing system.
func Process(c http.ResponseWriter, req *http.Request) {
	mainServer.Process(c, req)
}

// Run starts the web application and serves HTTP requests for the main server.
func Run(addr string) {
	mainServer.Run(addr)
}

// RunTLS starts the web application and serves HTTPS requests for the main server.
func RunTLS(addr string, config *tls.Config) {
	mainServer.RunTLS(addr, config)
}

// Close stops the main server.
func Close() {
	mainServer.Close()
}

// Get adds a handler for the 'GET' http method in the main server.
func Get(route string, handler interface{}) {
	mainServer.Get(route, handler)
}

// Post adds a handler for the 'POST' http method in the main server.
func Post(route string, handler interface{}) {
	mainServer.Post(route, handler)
}

// Post adds a handler for the 'POST' http method in the main server.
func Via(route string, handler interface{}) {
	mainServer.Get(route, handler)
	mainServer.Post(route, handler)
}

// Put adds a handler for the 'PUT' http method in the main server.
func Put(route string, handler interface{}) {
	mainServer.Put(route, handler)
}

// Delete adds a handler for the 'DELETE' http method in the main server.
func Delete(route string, handler interface{}) {
	mainServer.Delete(route, handler)
}

// Match adds a handler for an arbitrary http method in the main server.
func Match(method string, route string, handler interface{}) {
	mainServer.Match(route, method, handler)
}

// Adds a custom handler. Only for webserver mode. Will have no effect when running as FCGI or SCGI.
func Handler(route string, method string, httpHandler http.Handler) {
	mainServer.Handler(route, method, httpHandler)
}

// Default server
func App() *Server {
	return mainServer
}

var mainServer = NewServer()

// Run starts the web application and serves Tcp service for the main server.
func RunTcp(addr string) {
	mainTcp.Run(addr)
}

// Add a handler for tcp method in the main server.
func ViaTcp(route string, handler interface{}) {
	mainTcp.Via(route, handler)
}

// Add a plugin
func MiddlewareTcp(handler interface{}) {
	mainTcp.Middleware(handler)
}

// Default Tcp server
func AppTcp() *Tcp {
	return mainTcp
}

var mainTcp = NewTcp()

// -------------
// Run starts the web application and serves Tcp service for the main server.
func RunDuo(addr string) {
	mainDuo.Run(addr)
}

// Add a handler for tcp method in the main server.
func ViaDuo(route string, handler interface{}) {
	mainDuo.Via(route, handler)
}

// Add a plugin
func MiddlewareDuo(handler interface{}) {
	mainDuo.Middleware(handler)
}

// Default Duo server
func AppDuo() *Duo {
	return mainDuo
}

var mainDuo = NewDuo()
