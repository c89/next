// Package web is a lightweight web framework for Go. It's ideal for
// writing simple, performant backend web services.
package next

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// A Context object is created for every incoming HTTP request, and is
// passed to handlers as an optional first argument. It provides information
// about the request, including the http.Request object, the GET and POST params,
// and acts as a Writer for the response.
type Context struct {
	Request *http.Request
	Params  map[string]string
	Server  *Server
	http.ResponseWriter
}

// WriteString writes string data into the response object.
func (ctx *Context) WriteString(content string) {
	ctx.ResponseWriter.Write([]byte(content))
}

// WriteJSON writes json data into the response object.
func (ctx *Context) WriteJSON(code, msg string, data ...interface{}) {
	ctx.ContentType("json")

	json := NewJson()
	if ctx.Server.Config.Bool("debug.profiler") {
		json.Set("debug", ctx.Params)
	}

	json.Set("code", code)
	json.Set("msg", msg)

	if len(data) > 0 {
		json.Set("data", data[0])
	}
	out, _ := json.Encode()

	ctx.WriteString(string(out))
}

// Abort is a helper method that sends an HTTP header and an optional
// body. It is useful for returning 4xx or 5xx errors.
// Once it has been called, any return value from the handler will
// not be written to the response.
func (ctx *Context) Abort(status int, body string) {
	ctx.ResponseWriter.WriteHeader(status)
	ctx.ResponseWriter.Write([]byte(body))
}

// Notmodified writes a 304 HTTP response
func (ctx *Context) NotModified() {
	ctx.ResponseWriter.WriteHeader(304)
}

// NotFound writes a 404 HTTP response
func (ctx *Context) NotFound(message string) {
	ctx.ResponseWriter.WriteHeader(404)
	ctx.ResponseWriter.Write([]byte(message))
}

//Unauthorized writes a 401 HTTP response
func (ctx *Context) Unauthorized() {
	ctx.ResponseWriter.WriteHeader(401)
}

//Forbidden writes a 403 HTTP response
func (ctx *Context) Forbidden() {
	ctx.ResponseWriter.WriteHeader(403)
}

// ContentType sets the Content-Type header for an HTTP response.
// For example, ctx.ContentType("json") sets the content-type to "application/json"
// If the supplied value contains a slash (/) it is set as the Content-Type
// verbatim. The return value is the content type as it was
// set, or an empty string if none was found.
func (ctx *Context) ContentType(val string) string {
	var ctype string
	if strings.ContainsRune(val, '/') {
		ctype = val
	} else {
		if !strings.HasPrefix(val, ".") {
			val = "." + val
		}
		ctype = mime.TypeByExtension(val)
	}
	if ctype != "" {
		ctx.Header().Set("Content-Type", ctype)
	}
	return ctype
}

// SetHeader sets a response header. If `unique` is true, the current value
// of that header will be overwritten . If false, it will be appended.
func (ctx *Context) SetHeader(hdr string, val string, unique bool) {
	if unique {
		ctx.Header().Set(hdr, val)
	} else {
		ctx.Header().Add(hdr, val)
	}
}

// SetCookie adds a cookie header to the response.
func (ctx *Context) SetCookie(cookie *http.Cookie) {
	ctx.SetHeader("Set-Cookie", cookie.String(), false)
}

func (ctx *Context) GetCookieSig(key string, val []byte, timestamp string) string {
	hm := hmac.New(sha1.New, []byte(key))

	hm.Write(val)
	hm.Write([]byte(timestamp))

	hex := fmt.Sprintf("%02x", hm.Sum(nil))
	return hex
}

func (ctx *Context) SetSecureCookie(name string, val string, age int64) {
	//base64 encode the val
	if ctx.Server.Config.String("cookie.secret") == "" {
		ctx.Server.Logger.Println("Secret Key for secure cookies has not been set. Please assign a cookie secret to cookie.secret.")
		return
	}

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write([]byte(val))
	encoder.Close()
	vs := buf.String()
	vb := buf.Bytes()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sig := ctx.GetCookieSig(ctx.Server.Config.String("cookie.secret"), vb, timestamp)
	cookie := strings.Join([]string{vs, timestamp, sig}, "|")
	ctx.SetCookie(NewCookie(name, cookie, age))
}

func (ctx *Context) GetSecureCookie(name string) (string, bool) {
	for _, cookie := range ctx.Request.Cookies() {
		if cookie.Name != name {
			continue
		}

		parts := strings.SplitN(cookie.Value, "|", 3)

		val := parts[0]
		timestamp := parts[1]
		sig := parts[2]

		if ctx.GetCookieSig(ctx.Server.Config.String("cookie.secret"), []byte(val), timestamp) != sig {
			return "", false
		}

		ts, _ := strconv.ParseInt(timestamp, 0, 64)

		if time.Now().Unix()-31*86400 > ts {
			return "", false
		}

		buf := bytes.NewBufferString(val)
		encoder := base64.NewDecoder(base64.StdEncoding, buf)

		res, _ := ioutil.ReadAll(encoder)
		return string(res), true
	}
	return "", false
}

func (ctx *Context) ClientIp() (string, error) {
	ip, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("userip: %q is not IP:port", ctx.Request.RemoteAddr)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return "", fmt.Errorf("userip: %q is not IP:port", ctx.Request.RemoteAddr)
	}
	return userIP.String(), nil
}
