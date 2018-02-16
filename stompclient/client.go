package stompclient

import (
	"github.com/gmallard/stompngo"
	"errors"
	"fmt"
	"reflect"
	"github.com/go-stomp/stomp"
)

type MessageHandler func([]byte) (bool, error)

type Subscription struct {
	destination string
	s           *bool
	c           *Client
	handler     MessageHandler
}

type Client struct {
	conn                              *stompngo.Connection
	network, addr, username, password string
	running, connected                bool
	subs                              map[string]Subscription
}

func Create(network, addr, username, password string) (*Client, error) {
	c := new(Client)

	c.running = true

	c.network = network
	c.addr = addr
	c.username = username
	c.password = password

	return c, c.connect()
}

func (c *Client) ccc() bool {
	return reflect.ValueOf(c.conn).FieldByName("closed").Bool()
}

func (c *Client) connect() error {
	if c.conn != nil {
		c.conn.Disconnect()
	}


	stompngo.Connect()

	conn, err := stomp.Dial(c.network, c.addr,
		stomp.ConnOpt.Login(c.username, c.password))
	if err != nil {
		return err
	}

	c.conn = conn
	c.connected = true

	return nil
}

func (c *Client) Send(destination, contentType string, content []byte) error {
	if !c.running {
		return errors.New("client not running")
	}

	return c.conn.Send(destination, contentType, content)
}

func (c *Client) Subscribe(destination string, handler MessageHandler) (*Subscription, error) {
	if c.running == false {
		return nil, errors.New("client not running")
	}

	sub := &Subscription{destination: destination, c: c, handler: handler}
	defer c.doSubscribe(sub)
	return sub, nil
}

func (c *Client) doSubscribe(sub *Subscription) {
	s, err := c.conn.Subscribe(sub.destination, stomp.AckClient);
	if err != nil {
		fmt.Println("unable to subscribe", err)
		return
	}

	sub.s = s
	go c.doReceive(sub)
}

func (c *Client) doReceive(sub *Subscription) error {
	for {
		msg := <-sub.s.C
		if msg.Err != nil {
			return c.handlerError("error on receive msg from ", sub.destination, ":", msg.Err)
		}
		ack, err := sub.handler(msg.Body)

		if err != nil {
			fmt.Println("error hadling msg", err)
			continue
		}

		if (ack) {
			err = c.conn.Ack(msg)
			if err != nil {
				fmt.Println()
				return c.handlerError("error acking msg", err)
			}
		}
	}
}

func (c *Client) handlerError(a ...interface{}) error {
	fmt.Println(a)

	if c.running {
		c.connected = false
		c.connect()
	}

	return nil
}

func (c *Client) Close() error {
	c.running = false
	c.conn.Disconnect()
	return nil
}
