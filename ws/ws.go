package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chenqgp/abc"
	"github.com/gin-gonic/gin"
	"nhooyr.io/websocket"
)

var (
	maxConnect uint32 = 6000
	queue      uint32

	upgrader = websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	}

	maxMessageSize int64 = 3000
	readWait             = 15 * time.Second
	writeWait            = 15 * time.Second
	heartbeat            = 5 * time.Second
	limiterPeriod        = 20 * time.Second
	deadline             = 60 * 1000

	Message  = make(chan []byte)
	workers  = 3
	subToken = `:`
)

var Conn *connections

var closedchan = make(chan interface{})

type connections struct {
	c  map[int]*Server
	mu sync.Mutex
}

type Server struct {
	client    *websocket.Conn
	user      int
	keepalive chan interface{}
	done      atomic.Value
	mu        sync.Mutex
	msg       chan []byte
	subcribe  string
}

func init() {
	close(closedchan)
}

func (server *Server) cancel() {
	server.mu.Lock()
	d, _ := server.done.Load().(chan interface{})
	if d == nil {
		server.done.Store(closedchan)
	} else {
		//log.Println("close the done")
		close(d)
	}
	server.mu.Unlock()
}

func (server *Server) Done() <-chan interface{} {
	d := server.done.Load()
	if d != nil {
		return d.(chan interface{})
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	d = server.done.Load()
	if d == nil {
		//log.Println("make the done")
		d = make(chan interface{})
		server.done.Store(d)
	}
	return d.(chan interface{})
}

func (server *Server) read(ctx context.Context) {
	context.TODO()
	defer close(server.keepalive)
	defer close(server.msg)
	server.client.SetReadLimit(maxMessageSize)
	for {
		select {
		case <-server.Done():
			//log.Println("read done")
			break
		default:
		}

		//c.SetReadDeadline(time.Now().Add(readWait))
		_, msg, err := server.client.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusGoingAway ||
				websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusAbnormalClosure {
				//log.Println("read error", err)
			}
			select {
			case <-server.Done():
			default:
				server.cancel()
			}

			server.client.Close(websocket.StatusNormalClosure, "用户主动断开")
			Conn.remove(server)
			if old := atomic.SwapUint32(&queue,
				atomic.LoadUint32(&queue)-1); old == 0 {
				atomic.StoreUint32(&queue, 0)
			}
			// test log something..
			//log.Println(writeTimeout(ctx, writeWait, server.client, []byte("done")))

			d, _ := server.done.Load().(chan interface{})
			log.Println("current connect", len(Conn.c),
				"compare current queues", queue, d == closedchan)
			break
		}

		var s []string
		var subs string
		if string(msg) == "PING" {
			//fmt.Println(string(msg))
			server.keepalive <- true
		} else if err := json.Unmarshal(msg, &s); err == nil {
			//fmt.Println(string(msg))
			for _, sub := range s {
				subs = subs + sub + subToken
			}
			server.mu.Lock()
			server.subcribe = subs
			server.mu.Unlock()
		}
	}
}

func WsHandler(c *gin.Context) {
	if Conn == nil {
		Conn = &connections{
			c: make(map[int]*Server),
		}
	}

	Conn.Consumer(c)
}

func (conn *connections) Consumer(wr *gin.Context) {
	if queue >= maxConnect {
		log.Println("current connections are full", queue)
		return
	}
	token := abc.FetchFromToken(wr.Query("token"))
	if token.Expire < time.Now().Unix() {
		log.Println("have no permissions")
		return
	}

	c, err := websocket.Accept(wr.Writer, wr.Request, &upgrader)
	if err != nil {
		panic(err)
	}
	ctx := wr.Request.Context()

	server := &Server{
		client:    c,
		keepalive: make(chan interface{}, 1),
		msg:       make(chan []byte, 800), // 拟定订阅类目并发
		user:      token.Uid,
	}

	atomic.AddUint32(&queue, 1)

	conn.add(server)

	go server.read(ctx)
	for i := 0; i < workers; i++ {
		go server.write(ctx)
	}

	shutdown := func() {
		select {
		case <-server.Done():
			return
		default:
		}
		server.cancel()
		c.Close(websocket.StatusAbnormalClosure, "客户端断联")
		conn.remove(server)
		if old := atomic.SwapUint32(&queue,
			atomic.LoadUint32(&queue)-1); old == 0 {
			atomic.StoreUint32(&queue, 0)
		}
		// test log something..
		//log.Println(writeTimeout(ctx, writeWait, c, []byte("done")))

		d, _ := server.done.Load().(chan interface{})
		log.Println("curren connect", len(conn.c),
			"current queues", queue, d == closedchan)
	}

	// deadline.
	go func() {
		deadLine := time.NewTicker(limiterPeriod)
		//log.Println("deadline start", deadline, "ms")
		d, _ := time.ParseDuration(fmt.Sprintf("%dms", deadline))
		tEnd := time.Now().Add(d)
		for {
			select {
			case <-server.Done():
				//log.Println("deadline done before")
				return
			case <-deadLine.C:
				remain := tEnd.Sub(time.Now()).Seconds()
				//log.Println("deadline", remain)
				if remain <= 0.0 {
					//log.Println("deadline!")
					//shutdown()
					return
				}
			}
		}
	}()

	// heartbeat.
	go func() {
		ticker := time.NewTicker(heartbeat)
		limiter := time.NewTicker(limiterPeriod)
		go func() {
			for {
				select {
				case <-ticker.C:
					//log.Println("in")
					if _, ok := <-server.keepalive; ok {
						ticker.Reset(heartbeat)
						limiter.Reset(limiterPeriod)
						//log.Println("keepalive")
					}
				}
			}
		}()

		for {
			//log.Println("limiter period tick...")
			select {
			case <-limiter.C:
				log.Println("timeout")
				ticker.Stop()
				limiter.Stop()
				shutdown()
				return
			}
		}
	}()

	// test..
	//go func() {
	//	for {
	//		time.Sleep(5 * time.Second)
	//		Message <- []byte("writer:{data:" + abc.RandStr(10) + "}")
	//	}
	//}()

	for {
		select {
		case <-server.Done():
			return
		case msg := <-Message:
			conn.writeMessage(ctx, msg)
		}
	}

	log.Println("consumer shutdown")
	shutdown()
}

func (server *Server) Send(msg []byte) {
	select {
	case <-server.Done():
		return
	case server.msg <- msg:
	}
}

func Send(msg []byte) {
	go func() { Message <- msg }()
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}

func (server *Server) write(ctx context.Context) {
	for {
		select {
		case <-server.Done():
			return
		case msg := <-server.msg:
			if contains(server.subcribe, string(msg)[0:6]) {
				err := writeTimeout(ctx, writeWait, server.client, msg)
				if err != nil {
					log.Println("write error:", err)
				}
			}
		}
	}
}

func (conn *connections) writeMessage(ctx context.Context, message []byte) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	for _, server := range conn.c {
		select {
		case <-server.Done():
		default:
			server.msg <- message
		}
	}
}

func (conn *connections) add(server *Server) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.c[server.user] = server
}

func (conn *connections) remove(server *Server) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	delete(conn.c, server.user)
}

func contains(s, substr string) bool {
	n := len(substr)
	c0 := substr[0:5]
	c1 := substr[5:6]
	i := 0
	t := len(s) - n + 1
	for i < t {
		if s[i:i+5] == c0 && s[i+5:i+6] == c1 {
			return true
		}
		if s[i+5:i+6] != subToken {
			i++
		}
		i += 6
	}
	return false
}

func SingleBoardcast(uid int) (*Server, bool) {
	if Conn == nil {
		return nil, false
	}
	c, ok := Conn.c[uid]
	return c, ok
}
