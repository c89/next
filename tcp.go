package next

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"runtime"
)

type Tcp struct {
	Conn   map[int]*net.TCPConn
	Config *Config
	Logger *log.Logger
	routes *Routes
}

const (
	TcpHead       = 0xAA
	TcpTail       = 0x0A
	TcpMaxContent = 2 << 10
)

func NewTcp() *Tcp {
	tcp := &Tcp{
		Conn:   make(map[int]*net.TCPConn),
		Config: NewConfig(),
		Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
		routes: NewRoutes(),
	}

	// Load default config if exists
	file := "config.json"
	if fileExists(file) {
		tcp.Config.Load(file)
	}

	return tcp
}

// Pack is a utility function to read from the supplied Writer
// according to the Next protocol spec:
//
//    AA[x][x][x][x][x][x][x][x]55
//      |  (int32) || (binary)
//      |  4-byte  || N-byte
//      ------------------------...
//        size       data
func (t *Tcp) Pack(w io.Writer, data []byte) error {
	// Head
	_, err := w.Write([]byte{TcpHead})
	if err != nil {
		return errors.New("write head fail")
	}

	// Size
	err = binary.Write(w, binary.LittleEndian, int32(len(data))+6)
	if err != nil {
		return errors.New("write size fail")
	}

	// Data
	_, err = w.Write(data)
	if err != nil {
		return errors.New("write data fail")
	}

	// Tail
	_, err = w.Write([]byte{TcpTail})
	if err != nil {
		return errors.New("write tail fail")
	}

	return nil
}

// Unpack is a utility function to read from the supplied Reader
// according to the Next protocol spec:
//
//    AA[x][x][x][x][x][x][x][x]55
//      |  (int32) || (binary)
//      |  4-byte  || N-byte
//      ------------------------...
//        size       data
func (t *Tcp) Unpack(r io.Reader) ([]byte, error) {
	// Check head
	head := make([]byte, 1)
	_, err := r.Read(head)
	if err != nil {
		return nil, err
	}
	if int32(head[0]) != TcpHead {
		return nil, errors.New("data head is error")
	}

	// message size
	var size int32
	err = binary.Read(r, binary.LittleEndian, &size)
	if err != nil {
		return nil, err
	}
	size = size - 6
	if size <= 0 || size > TcpMaxContent {
		return nil, errors.New("data size is error")
	}

	// message binary data
	buf := make([]byte, size)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	// Check head
	tail := make([]byte, 1)
	_, err = r.Read(tail)
	if err != nil {
		return nil, err
	}
	if int32(tail[0]) != TcpTail {
		return nil, errors.New("data tail is error")
	}

	return buf, nil
}

// Get the integer Unix file descriptor referencing the open file
func (t *Tcp) Fd(conn *net.TCPConn) (int, error) {
	f, err := conn.File()
	if err != nil {
		return -1, err
	}

	return int(f.Fd()), nil
}

func (t *Tcp) Pipe(conn *net.TCPConn) {
	fd, err := t.Fd(conn)
	if err != nil {
		t.Logger.Printf("connect err: %s\n", err)
		return
	}
	defer func() {
		t.Logger.Printf("disconnected: %s\n", conn.RemoteAddr().String())
		conn.Close()
		delete(t.Conn, fd)
	}()

	// Save in map
	t.Conn[fd] = conn

	// Read data
	reader := bufio.NewReader(conn)
	for {
		body, err := t.Unpack(reader)
		t.Logger.Printf("%x", body)
		if err != nil {
			if err == io.EOF {
				return
			}
			t.Logger.Print(err)
		}

		// Filter heart pack
		if string(body) != "hello" {
			t.handler(conn, body)
		}
	}
}

func (t *Tcp) Run(addr string) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	tcpListener, _ := net.ListenTCP("tcp", tcpAddr)
	defer tcpListener.Close()

	t.Logger.Printf("next tcp serving %s\n", addr)
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
func (t *Tcp) Via(route string, handler interface{}) {
	t.routes.Add(route, "VIA", handler)
}

// safelyCall invokes `function` in recover block
func (t *Tcp) safelyCall(function reflect.Value, args []reflect.Value) (resp []reflect.Value, e interface{}) {
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

func (t *Tcp) handler(conn *net.TCPConn, body []byte) {
	ctx := TcpContext{
		Params: make(map[string]string),
		Tcp:    t,
		conn:   conn,
	}

	// Read json body
	json := NewJson()
	json.Load(body)
	requestPath := json.Get("method").MustString()

	route := t.routes.Match(requestPath, "VIA")
	if route == nil {
		ctx.WriteJSON("404", "Method not found")
		return
	}
	cr := route.cr

	var args []reflect.Value
	handlerType := route.handler.Type()
	if requiresContext(handlerType) {
		args = append(args, reflect.ValueOf(&ctx))
	}

	match := cr.FindStringSubmatch(requestPath)
	for _, arg := range match[1:] {
		args = append(args, reflect.ValueOf(arg))
	}

	_, err := t.safelyCall(route.handler, args)
	if err != nil {
		//there was an error or panic while calling the handler
		ctx.WriteJSON("500", "Server Error")
	}
}

// --------
// Tcp Context
// --------
type TcpContext struct {
	Params map[string]string
	Tcp    *Tcp
	conn   *net.TCPConn
}

// WriteJSON writes json data into the response object.
func (ctx *TcpContext) WriteJSON(code, msg string, data ...interface{}) {
	json := NewJson()
	if ctx.Tcp.Config.Bool("debug.profiler") {
		json.Set("debug", ctx.Params)
	}

	json.Set("code", code)
	json.Set("msg", msg)

	if len(data) > 0 {
		json.Set("data", data[0])
	}
	out, _ := json.Encode()

	ctx.Tcp.Pack(ctx.conn, out)
}
