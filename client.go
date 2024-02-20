package OpenAPI

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"io"
	"time"

	"github.com/asaskevich/EventBus"
)

type IClient interface {
	Connect() error
	Disconnect()
	SendMessage(m []byte) error
	// On "message" | "end", EventHandler
	On(string, EventHandler)
	OnGeneric(string, EventHandlerGeneric)
	// Off "message" | "end", EventHandler
	Off(string, EventHandler)
}

type ClientConfig struct {
	Address      string
	ClientID     string
	ClientSecret string
	CertFile     string
	KeyFile      string
}

type Client struct {
	Address      string
	ClientID     string
	ClientSecret string
	CertFile     string
	KeyFile      string
	connected    bool
	reader       *bufio.Reader
	bus          EventBus.Bus
	conn         *tls.Conn
}

type EventHandler func([]byte)

type EventHandlerGeneric func(interface{})

func NewClient(config ClientConfig) IClient {
	client := &Client{
		Address:      config.Address,
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		CertFile:     config.CertFile,
		KeyFile:      config.KeyFile,
	}
	client.bus = EventBus.New()
	return client
}

func (c *Client) Connect() error {
	cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return err
	}
	c.conn, err = tls.Dial("tcp", c.Address, &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	c.connected = true
	c.reader = bufio.NewReader(c.conn)
	go c.readMessages()
	go c.sendHearBeats()
	return nil
}

func (c *Client) Disconnect() {
	c.conn.Close()
	c.connected = false
}

func (c *Client) SendMessage(m []byte) error {
	size := ToByteArray(len(m))
	Reverse(size)
	//@todo: Add error handling
	//@todo: Set read/write deadlines
	c.conn.Write(size)
	c.conn.Write(m)
	return nil
}

func (c *Client) readMessage() ([]byte, error) {
	// read message length
	firstFourBytes := make([]byte, 4)
	_, err := c.reader.Read(firstFourBytes)
	if err != nil {
		return nil, err
	}
	//Reverse(firstFourBytes)
	messageLen := binary.BigEndian.Uint32(firstFourBytes)
	// read message content
	packet := make([]byte, messageLen)
	_, err = io.ReadFull(c.reader, packet)
	if err != nil {
		return nil, err
	}
	return packet, nil
}

func (c *Client) readMessages() {
	for {
		message, err := c.readMessage()
		//There is a general lack of official support
		// to detect whether a server closed a TCP connection
		// https://stackoverflow.com/questions/12741386/how-to-know-tcp-connection-is-closed-in-net-package/12741495#12741495
		if err != nil {
			c.connected = false
			c.bus.Publish("end", "Connection ended")
			break
		}
		c.bus.Publish("message", message)
	}
}

func (c *Client) sendHearBeats() {
	for range time.Tick(time.Second * 30) {
		if c.connected {
			// A shortcut for &model.ProtoHeartbeatEvent{}
			c.SendMessage([]byte{8, 51, 18, 0})
		}
	}
}

func (c *Client) On(event string, cb EventHandler) {
	c.bus.Subscribe(event, cb)
}

func (c *Client) Off(event string, cb EventHandler) {
	c.bus.Unsubscribe(event, cb)
}

func (c *Client) OnGeneric(event string, cb EventHandlerGeneric) {
	c.bus.Subscribe(event, cb)
}
