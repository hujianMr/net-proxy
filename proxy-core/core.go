package proxy_core

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"io"
	"log"
	"net"
	"runtime/debug"
	"sync"
	"time"
)

type Request struct {
	Conn net.Conn
	Buff []byte
}

type Server struct {
	Server    net.Listener
	V         int
	Client    net.Conn
	ProxyPort string
}

func (s *Server) IncrCycle(client net.Conn) Server {
	s.V++
	s.Client = client
	return Server{
		Server:    s.Server,
		V:         s.V,
		Client:    s.Client,
		ProxyPort: s.ProxyPort,
	}
}

func (s *Server) Expire(V int) bool {
	if s.V < V {
		return true
	}
	return false
}

// 关闭连接
func Close(cons ...net.Conn) {
	for _, conn := range cons {
		_ = conn.Close()
	}
}

func ListenServer(proxyPort string) (net.Listener, error) {
	address := "0.0.0.0:" + proxyPort
	log.Println("listen addr ：" + address)
	server, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return server, nil
}

func IoBind(dst io.ReadWriter, src io.ReadWriter, fn func(isSrcErr bool, err error), cfn func(count int, isPositive bool), bytesPreSec float64) {
	var one = &sync.Once{}
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("IoBind crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
			}
		}()
		var err error
		var isSrcErr bool
		if bytesPreSec > 0 {
			newreader := NewReader(src)
			newreader.SetRateLimit(bytesPreSec)
			_, isSrcErr, err = ioCopy(dst, newreader, func(c int) {
				cfn(c, false)
			})

		} else {
			_, isSrcErr, err = ioCopy(dst, src, func(c int) {
				cfn(c, false)
			})
		}
		if err != nil {
			one.Do(func() {
				fn(isSrcErr, err)
			})
		}
	}()
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("IoBind crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
			}
		}()
		var err error
		var isSrcErr bool
		if bytesPreSec > 0 {
			newReader := NewReader(dst)
			newReader.SetRateLimit(bytesPreSec)
			_, isSrcErr, err = ioCopy(src, newReader, func(c int) {
				cfn(c, true)
			})
		} else {
			_, isSrcErr, err = ioCopy(src, dst, func(c int) {
				cfn(c, true)
			})
		}
		if err != nil {
			one.Do(func() {
				fn(isSrcErr, err)
			})
		}
	}()
}
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:   r,
		ctx: context.Background(),
	}
}
func ioCopy(dst io.Writer, src io.Reader, fn ...func(count int)) (written int64, isSrcErr bool, err error) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
				if len(fn) == 1 {
					fn[0](nw)
				}
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			err = er
			isSrcErr = true
			break
		}
	}
	return written, isSrcErr, err
}

func ProxyIoBind(inConn *net.Conn, outConn net.Conn) {
	inAddr := (*inConn).RemoteAddr().String()
	inLocalAddr := (*inConn).LocalAddr().String()
	outAddr := outConn.RemoteAddr().String()
	outLocalAddr := outConn.LocalAddr().String()
	IoBind(*inConn, outConn, func(isSrcErr bool, err error) {
		log.Printf("conn %s - %s - %s -%s released", inAddr, inLocalAddr, outLocalAddr, outAddr)
		CloseConn(inConn)
		CloseConn(&outConn)
	}, func(n int, d bool) {}, 0)
}

func CloseConn(conn *net.Conn) {
	if conn != nil && *conn != nil {
		(*conn).SetDeadline(time.Now().Add(time.Millisecond))
		(*conn).Close()
	}
}

func ProxySwap(proxyConn net.Conn, client net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go ConnCopy(proxyConn, client, &wg)
	go ConnCopy(client, proxyConn, &wg)
	wg.Wait()
	log.Println("conn1 = [" + proxyConn.LocalAddr().String() + "], conn2 = [" + client.RemoteAddr().String() + "] iocopy读写完成")
}
func ConnCopy(conn1 net.Conn, conn2 net.Conn, wg *sync.WaitGroup) {
	_, err := io.Copy(conn1, conn2)
	if err != nil {
		log.Println("conn1 = ["+conn1.LocalAddr().String()+"], conn2 = ["+conn2.RemoteAddr().String()+"] iocopy失败", err)
	}
	log.Println("[←]", "close the connect at local:["+conn1.LocalAddr().String()+"] and remote:["+conn1.RemoteAddr().String()+"]")
	conn1.Close()
	wg.Done()
}

func PrintWelcome() {
	fmt.Println("+----------------------------------------------------------------+")
	fmt.Println("| welcome use ho-huj-net-proxy Version1.0                        |")
	fmt.Println("| author Ruchsky at 2019-11-14                                   |")
	fmt.Println("| gitee home page ->   https://gitee.com/ruchsky                 |")
	fmt.Println("| github home page ->  https://github.com/hujianMr               |")
	fmt.Println("+----------------------------------------------------------------+")
	fmt.Print("start..")
	i := 0
	for {
		fmt.Print(">>>>")
		i++
		time.Sleep(time.Second)
		if i >= 10 {
			break
		}
	}
	fmt.Println()
	fmt.Println("start success")
	time.Sleep(time.Second)
}

const burstLimit = 1000 * 1000 * 1000

type Reader struct {
	r       io.Reader
	limiter *rate.Limiter
	ctx     context.Context
}

type Writer struct {
	w       io.Writer
	limiter *rate.Limiter
	ctx     context.Context
}

// NewReaderWithContext returns a reader that implements io.Reader with rate limiting.
func NewReaderWithContext(r io.Reader, ctx context.Context) *Reader {
	return &Reader{
		r:   r,
		ctx: ctx,
	}
}

// NewWriter returns a writer that implements io.Writer with rate limiting.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:   w,
		ctx: context.Background(),
	}
}

// NewWriterWithContext returns a writer that implements io.Writer with rate limiting.
func NewWriterWithContext(w io.Writer, ctx context.Context) *Writer {
	return &Writer{
		w:   w,
		ctx: ctx,
	}
}

// SetRateLimit sets rate limit (bytes/sec) to the reader.
func (s *Reader) SetRateLimit(bytesPerSec float64) {
	s.limiter = rate.NewLimiter(rate.Limit(bytesPerSec), burstLimit)
	s.limiter.AllowN(time.Now(), burstLimit) // spend initial burst
}

// Read reads bytes into p.
func (s *Reader) Read(p []byte) (int, error) {
	if s.limiter == nil {
		return s.r.Read(p)
	}
	n, err := s.r.Read(p)
	if err != nil {
		return n, err
	}
	if err := s.limiter.WaitN(s.ctx, n); err != nil {
		return n, err
	}
	return n, nil
}

// SetRateLimit sets rate limit (bytes/sec) to the writer.
func (s *Writer) SetRateLimit(bytesPerSec float64) {
	s.limiter = rate.NewLimiter(rate.Limit(bytesPerSec), burstLimit)
	s.limiter.AllowN(time.Now(), burstLimit) // spend initial burst
}

// Write writes bytes from p.
func (s *Writer) Write(p []byte) (int, error) {
	if s.limiter == nil {
		return s.w.Write(p)
	}
	n, err := s.w.Write(p)
	if err != nil {
		return n, err
	}
	if err := s.limiter.WaitN(s.ctx, n); err != nil {
		return n, err
	}
	return n, err
}
