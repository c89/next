package next

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"runtime"
	"time"
)

type Duo struct {
	Conn       map[string]*net.TCPConn
	Config     *Config
	Logger     *log.Logger
	routes     *Routes
	middleware []reflect.Value
}

const (
	DuoHead       = 0xEF
	DuoSec        = 0x01
	DuoMaxContent = 2 << 10
)

func NewDuo() *Duo {
	duo := &Duo{
		Conn:       make(map[string]*net.TCPConn),
		Config:     NewConfig(),
		Logger:     log.New(os.Stdout, "", log.Ldate|log.Ltime),
		routes:     NewRoutes(),
		middleware: make([]reflect.Value, 0),
	}

	// Load default config if exists
	file := "config.json"
	if fileExists(file) {
		duo.Config.Load(file)
	}

	return duo
}

// Pack is a utility function to read from the supplied Writer
// according to the Next protocol spec:
//
//    EF01[x][x][x][x][x][x][x][x][x]
//        |  | (binary)          |
//        |1-byte                | 1-byte
//      ---------------------------
//    head size       data        crc
func (t *Duo) Pack(w io.Writer, data []byte) error {
	// Head
	out := []byte{DuoHead, DuoSec}
	// Size
	out = append(out, byte(len(data)+2))
	// Data
	out = append(out, data...)
	// Tail
	out = append(out, t.crc(data))
	if _, err := w.Write(out); err != nil {
		return errors.New("write fail")
	}

	return nil
}

// CRC data code
// 100-(EF+01+X)%FF
func (t *Duo) crc(data []byte) byte {
	sum := DuoHead + DuoSec + len(data) + 2
	for _, d := range data {
		sum += int(d)
	}

	return byte(0x100 - sum%0xFF)
}

// Unpack is a utility function to read from the supplied Reader
// according to the Next protocol spec:
//
//    EF01[x][x][x][x][x][x][x][x][x]
//        |  | (binary)          |
//        |1-byte                | 1-byte
//      ---------------------------
//    head size       data        crc
func (t *Duo) Unpack(r io.Reader) ([]byte, error) {
	// Check head
	head := make([]byte, 2)
	_, err := r.Read(head)
	if err != nil {
		return nil, err
	}
	if int32(head[0]) != DuoHead && int32(head[1]) != DuoSec {
		return nil, errors.New("data head is error")
	}

	// message size
	s := make([]byte, 1)
	_, err = r.Read(s)
	if err != nil {
		return nil, err
	}
	size := int32(uint32(s[0]))
	size = size - 2
	if size <= 0 || size > TcpMaxContent {
		return nil, errors.New("data size is error")
	}

	// message binary data
	buf := make([]byte, size)
	_, err = r.Read(buf)
	if err != nil {
		return nil, err
	}

	// Check head
	tail := make([]byte, 1)
	_, err = r.Read(tail)
	if err != nil {
		return nil, err
	}
	if tail[0] != t.crc(buf) {
		// TODO
		//		return nil, errors.New("data tail is error")
	}

	//t.Logger.Printf("Head: %v, size: %v, buf: %v, Tail: %v", head, size, buf, tail)

	return buf, nil
}

func (t *Duo) handler(conn *net.TCPConn, body []byte) {
	requestPath := "device" // Hack for device sever

	ctx := DuoContext{
		Method: body[0],
		Params: body,
		Duo:    t,
		Fd:     t.Fd(conn),
		conn:   conn,
	}
	tm := time.Now().UTC()
	defer t.logRequest(ctx, tm)

	route := t.routes.Match(requestPath, "VIA")
	if route == nil {
		//ctx.WriteJSON("404", "request method not found")
		return
	}

	cr := route.cr

	var args []reflect.Value
	handlerType := route.handler.Type()
	if requiresDuoContext(handlerType) {
		args = append(args, reflect.ValueOf(&ctx))
	}

	match := cr.FindStringSubmatch(requestPath)
	for _, arg := range match[1:] {
		args = append(args, reflect.ValueOf(arg))
	}

	t.safelyCall(route.handler, args)
}

// Get the integer Unix file descriptor referencing the open file
func (t *Duo) Fd(conn *net.TCPConn) string {
	return conn.RemoteAddr().String()
}

func (t *Duo) Pipe(conn *net.TCPConn) {
	fd := t.Fd(conn)
	defer func() {
		t.Logger.Printf("disconnected: %s\n", fd)
		conn.Close()
		delete(t.Conn, fd)
	}()

	// Save in map
	t.Conn[fd] = conn

	// Read data
	reader := bufio.NewReader(conn)
	for {
		body, err := t.Unpack(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				return
			}
			t.Logger.Print(err)
		}

		conn.SetReadDeadline(time.Now().Add(20 * time.Second))

		if len(body) == 0 {
			continue
		}

		t.handler(conn, body)
		reader.Reset(conn)
	}
}

func (t *Duo) Run(addr string) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	tcpListener, _ := net.ListenTCP("tcp", tcpAddr)
	defer tcpListener.Close()

	t.Logger.Printf("next duo serving %s\n", addr)

	// Run middleware
	go func() {
		for _, v := range t.middleware {
			var args []reflect.Value
			args = append(args, reflect.ValueOf(t))
			t.safelyCall(v, args)
		}
	}()

	// Run tcp listener
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			continue
		}

		t.Logger.Printf("connected: %s\n", tcpConn.RemoteAddr().String())
		go t.Pipe(tcpConn)
	}
}

// Post adds a handler for the 'Via' TCP method for tcp.
func (t *Duo) Middleware(handler interface{}) {
	switch handler.(type) {
	case reflect.Value:
		fv := handler.(reflect.Value)
		t.middleware = append(t.middleware, fv)
	default:
		fv := reflect.ValueOf(handler)
		t.middleware = append(t.middleware, fv)
	}
}

// Post adds a handler for the 'Via' TCP method for tcp.
func (t *Duo) Via(route string, handler interface{}) {
	t.routes.Add(route, "VIA", handler)
}

func (t *Duo) Write(conn *net.TCPConn, code byte, data ...[]byte) {
	out := make([]byte, 0)

	out = append(out, code)
	if len(data) > 0 {
		out = append(out, data[0]...)
	}

	t.Pack(conn, out)
}

// safelyCall invokes `function` in recover block
func (t *Duo) safelyCall(function reflect.Value, args []reflect.Value) (resp []reflect.Value, e interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if t.Config.Bool("panic") {
				// go back to panic
				panic(err)
			} else {
				e = err
				resp = nil
				t.Logger.Println("Handler crashed with error", err)
				for i := 1; ; i += 1 {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					t.Logger.Println(file, line)
				}
			}
		}
	}()
	return function.Call(args), nil
}

func (t *Duo) logRequest(ctx DuoContext, sTime time.Time) {
	//log the request
	var logEntry bytes.Buffer

	duration := time.Now().Sub(sTime)

	fmt.Fprintf(&logEntry, "%s - \033[32;1m TCP %x\033[0m - %v", ctx.Fd, ctx.Method, duration)

	if len(ctx.Params) > 0 {
		fmt.Fprintf(&logEntry, " - \033[37;1mParams: %v\033[0m\n", ctx.Params)
	}

	ctx.Duo.Logger.Print(logEntry.String())
}

// requiresContext determines whether 'handlerType' contains
// an argument to 'web.Ctx' as its first argument
func requiresDuoContext(handlerType reflect.Type) bool {
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
	if a0.Elem() == reflect.TypeOf(DuoContext{}) {
		return true
	}

	return false
}

// --------
// Duo Context
// --------
type DuoContext struct {
	Method byte
	Params []byte
	Duo    *Duo
	Fd     string
	conn   *net.TCPConn
}

// Writes data into the response object.
func (ctx *DuoContext) Write(out []byte) {
	ctx.Duo.Pack(ctx.conn, out)
}
