package fasthttp

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"
)

// workerPool serves incoming connections via a pool of workers
// in FILO order, i.e. the most recently stopped worker will serve the next
// incoming connection.
//
// Such a scheme keeps CPU caches hot (in theory).
type workerPool struct {
	// Function for serving server connections.
	// It must leave c unclosed.
	WorkerFunc ServeHandler

	MaxWorkersCount int

	LogAllErrors bool

	MaxIdleWorkerDuration time.Duration

	Logger Logger

	lock         sync.Mutex
	workersCount int
	mustStop     bool

	ready []*workerChan

	stopCh chan struct{}

	workerChanPool sync.Pool

	connState func(net.Conn, ConnState)
}

type workerChan struct {
	lastUseTime time.Time
	ch          chan net.Conn
}

func (wp *workerPool) Start() {
	if wp.stopCh != nil {
		panic("BUG: workerPool already started")
	}
	wp.stopCh = make(chan struct{})
	stopCh := wp.stopCh
	fmt.Println("WorkerPool Start")
	go func() {
		var scratch []*workerChan
		for {
			wp.clean(&scratch)
			select {
			case <-stopCh:
				fmt.Println("WorkerPool Stop")
				return
			default:
				time.Sleep(wp.getMaxIdleWorkerDuration())
				fmt.Println("WorkerPool Sleep,getMaxIdleWorkerDuration", wp.getMaxIdleWorkerDuration())
			}
		}
	}()
}

func (wp *workerPool) Stop() {
	if wp.stopCh == nil {
		panic("BUG: workerPool wasn't started")
	}
	close(wp.stopCh)
	wp.stopCh = nil

	// Stop all the workers waiting for incoming connections.
	// Do not wait for busy workers - they will stop after
	// serving the connection and noticing wp.mustStop = true.
	wp.lock.Lock()
	ready := wp.ready
	for i, ch := range ready {
		ch.ch <- nil
		ready[i] = nil
	}
	wp.ready = ready[:0]
	wp.mustStop = true
	wp.lock.Unlock()
}

func (wp *workerPool) getMaxIdleWorkerDuration() time.Duration {
	if wp.MaxIdleWorkerDuration <= 0 {
		return 10 * time.Second
	}
	return wp.MaxIdleWorkerDuration
}

func (wp *workerPool) clean(scratch *[]*workerChan) {
	maxIdleWorkerDuration := wp.getMaxIdleWorkerDuration()

	// Clean least recently used workers if they didn't serve connections
	// for more than maxIdleWorkerDuration.
	currentTime := time.Now()

	wp.lock.Lock()
	ready := wp.ready
	n := len(ready)
	i := 0
	for i < n && currentTime.Sub(ready[i].lastUseTime) > maxIdleWorkerDuration {
		i++
	}
	*scratch = append((*scratch)[:0], ready[:i]...)
	if i > 0 {
		m := copy(ready, ready[i:])
		for i = m; i < n; i++ {
			ready[i] = nil
		}
		wp.ready = ready[:m]
	}
	wp.lock.Unlock()

	// Notify obsolete workers to stop.
	// This notification must be outside the wp.lock, since ch.ch
	// may be blocking and may consume a lot of time if many workers
	// are located on non-local CPUs.
	tmp := *scratch
	for i, ch := range tmp {
		ch.ch <- nil
		tmp[i] = nil
	}
}

func (wp *workerPool) Serve(c net.Conn) bool {
	/**
	 *TODO::3，为当前连接找一个空闲的Channel，一开始一个都没有的时候，会创建一个channel，并创建一个协程监听此连接channel上的事件。此处关联TODO::6
	 *处里请求都在这个方法里面
	 */
	ch := wp.getCh()
	if ch == nil {
		return false
	}

	//TODO::6把请求放进该连接的请求队列中,上文监听的channel将会获取到
	ch.ch <- c
	return true
}

var workerChanCap = func() int {
	// Use blocking workerChan if GOMAXPROCS=1.
	// This immediately switches Serve to WorkerFunc, which results
	// in higher performance (under go1.5 at least).
	if runtime.GOMAXPROCS(0) == 1 {
		return 0
	}

	// Use non-blocking workerChan if GOMAXPROCS>1,
	// since otherwise the Serve caller (Acceptor) may lag accepting
	// new connections if WorkerFunc is CPU-bound.
	return 1
}()

/**
新建连接回走这里
 */
func (wp *workerPool) getCh() *workerChan {

	fmt.Println("getCh Get WorkerChan From workerPool")

	var ch *workerChan
	createWorker := false

	wp.lock.Lock()
	ready := wp.ready
	n := len(ready) - 1
	if n < 0 {
		if wp.workersCount < wp.MaxWorkersCount {
			createWorker = true
			wp.workersCount++
		}
	} else {
		ch = ready[n]
		ready[n] = nil
		wp.ready = ready[:n]
	}
	wp.lock.Unlock()

	/**
		默认是没有通道的，如果没有通道，则创建一个
	 */
	if ch == nil {
		if !createWorker {
			return nil
		}
		vch := wp.workerChanPool.Get()
		if vch == nil {
			//TODO::4，创建一个workerChan,放在chanal池中
			vch = &workerChan{
				ch: make(chan net.Conn, workerChanCap),
			}
		}
		ch = vch.(*workerChan)
		go func() {
			fmt.Println("每个连接开一个协程")
			/**
			 *TODO::5,处理请求，每个请求一个协程
			 */
			wp.workerFunc(ch)
			wp.workerChanPool.Put(vch)
		}()
	}
	return ch
}

func (wp *workerPool) release(ch *workerChan) bool {
	ch.lastUseTime = time.Now()
	wp.lock.Lock()
	if wp.mustStop {
		wp.lock.Unlock()
		return false
	}
	wp.ready = append(wp.ready, ch)
	wp.lock.Unlock()
	return true
}

func (wp *workerPool) workerFunc(ch *workerChan) {
	var c net.Conn

	fmt.Println("workerFunc Start,循环监听当前连接发送过来的请求")
	var err error
	for c = range ch.ch {
		fmt.Println("============获取到一个新的请求workerFunc=============")
		if c == nil {
			break
		}

		/*****
		TODO::7 wp.WorkerFunc方法，是workerPool中的WorkerFunc方法，此方法在Server初始化时指向了Server的serveConn方法
		func (s *Server) serveConn(c net.Conn) error {

		wp := &workerPool{
			//WorkerFunc-->Server.serveConn
			WorkerFunc:      s.serveConn,
			MaxWorkersCount: maxWorkersCount,
			LogAllErrors:    s.LogAllErrors,
			Logger:          s.logger(),
			connState:       s.setState,
		}
		***/
		if err = wp.WorkerFunc(c); err != nil && err != errHijacked {
			errStr := err.Error()
			if wp.LogAllErrors || !(strings.Contains(errStr, "broken pipe") ||
				strings.Contains(errStr, "reset by peer") ||
				strings.Contains(errStr, "request headers: small read buffer") ||
				strings.Contains(errStr, "i/o timeout")) {
				wp.Logger.Printf("error when serving connection %q<->%q: %s", c.LocalAddr(), c.RemoteAddr(), err)
			}
		}
		if err == errHijacked {
			wp.connState(c, StateHijacked)
		} else {
			c.Close()
			wp.connState(c, StateClosed)
		}
		c = nil

		if !wp.release(ch) {
			break
		}
	}

	wp.lock.Lock()
	wp.workersCount--
	wp.lock.Unlock()
}
