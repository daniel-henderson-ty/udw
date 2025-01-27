package udwChan

import (
	"github.com/tachyon-protocol/udw/udwErr"
	"strings"
	"sync"
)

type Chan struct {
	ch          chan interface{}
	isClosed    bool
	lock        sync.RWMutex
	closeSignal chan struct{}
}

func MakeChan(bufferSize int) *Chan {
	if bufferSize < 0 {
		return nil
	}
	sc := &Chan{
		ch:          make(chan interface{}, bufferSize),
		closeSignal: make(chan struct{}),
	}
	return sc
}

func (c *Chan) Send(data interface{}) (isClose bool) {
	c.lock.RLock()
	isClose = c.isClosed
	if isClose {
		c.lock.RUnlock()
		return true
	}
	select {
	case c.ch <- data:
	case <-c.closeSignal:
	}
	c.lock.RUnlock()
	return
}

func (c *Chan) SendIfEmpty(data interface{}) (isClose bool, isSuccess bool) {
	c.lock.RLock()
	isClose = c.isClosed
	if isClose {
		c.lock.RUnlock()
		return true, false
	}
	select {
	case c.ch <- data:
		isSuccess = true
	case <-c.closeSignal:
		isClose = true
	default:
	}
	c.lock.RUnlock()
	return
}

func (c *Chan) GetReceiveCh() <-chan interface{} {
	return c.ch
}

func (c *Chan) Receive() (data interface{}, isClose bool) {
	c.lock.RLock()
	if c.isClosed {
		c.lock.RUnlock()
		return nil, true
	}
	c.lock.RUnlock()
	i, ok := <-c.ch
	return i, !ok
}

func (c *Chan) Close() {

	err := udwErr.PanicToError(func() {
		close(c.closeSignal)
	})
	if err != nil {
		if !strings.Contains(err.Error(), "close of closed channel") {
			panic("n2mu5ht9p8 " + err.Error())
		}
	}
	c.lock.Lock()
	if c.isClosed {
		c.lock.Unlock()
		return
	}
	c.isClosed = true
	close(c.ch)
	c.lock.Unlock()
}

func (c *Chan) IsClosed() bool {
	c.lock.RLock()
	isClosed := c.isClosed
	c.lock.RUnlock()
	return isClosed
}
