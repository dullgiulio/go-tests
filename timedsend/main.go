package main

import (
	"fmt"
	"log"
	"time"
)

type Client struct {
	msgs    chan string
	deliver chan string
	errs    chan error
	done    chan struct{}
	senders int
}

func NewClient(senders int) *Client {
	c := &Client{
		senders: senders,
		msgs:    make(chan string),
		deliver: make(chan string),
		errs:    make(chan error, senders),      // Buffering optional
		done:    make(chan struct{}, senders+2), // Buffering optional
	}
	for i := 0; i < senders; i++ {
		go c.deliverMsgs()
	}
	go c.sendTimeout()
	go c.readErrors()
	return c
}

func (c *Client) deliverMsgs() {
	for msg := range c.deliver {
		// XXX: Real delivery operation that might take long
		<-time.After(time.Second * 1)
		fmt.Printf("%s: Message delivered\n", msg)
		// XXX: Failure here sends on c.errs and just continues
	}
	c.done <- struct{}{}
}

func (c *Client) sendTimeout() {
	for msg := range c.msgs {
		select {
		case c.deliver <- msg:
		case <-time.After(time.Second * 1):
			c.errs <- fmt.Errorf("%s: send timed out", msg)
		}
	}
	c.done <- struct{}{}
}

func (c *Client) readErrors() {
	for err := range c.errs {
		log.Print(err)
	}
	c.done <- struct{}{}
}

func (c *Client) Send(msg string) {
	c.msgs <- msg
}

func (c *Client) Close() {
	close(c.msgs)
	<-c.done
	close(c.errs)
	<-c.done
	close(c.deliver)
	for i := 0; i < c.senders; i++ {
		<-c.done
	}
}

func main() {
	c := NewClient(3)
	for i := 0; i < 6; i++ {
		c.Send(fmt.Sprintf("Msg #%d", i))
	}
	c.Close()
}
