package ami

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"strings"
	"sync"
)

// Socket holds the socket client connection data.
type Socket struct {
	conn     net.Conn
	incoming chan string
	shutdown chan struct{}
	wg       sync.WaitGroup
}

// NewSocket provides a new socket client, connecting to a tcp server.
func NewSocket(address string) (*Socket, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	s := &Socket{
		conn:     conn,
		incoming: make(chan string, 32),
		shutdown: make(chan struct{}),
	}
	s.wg.Add(1)
	go s.run(conn)
	return s, nil
}

// Connected returns the socket status, true for connected,
// false for disconnected.
func (s *Socket) Connected() bool {
	if s.conn == nil {
		return false
	}
	return true
}

// Close closes socket connection.
func (s *Socket) Close() error {
	close(s.shutdown)
	s.wg.Wait()

	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// Send sends data to socket using fprintf format.
func (s *Socket) Send(message string) error {
	_, err := s.conn.Write([]byte(message))
	return err
}

// Recv receives a string from socket server.
func (s *Socket) Recv() (string, error) {
	var buffer bytes.Buffer
	for {
		select {
		case msg := <-s.incoming:
			buffer.WriteString(msg)
			if strings.HasSuffix(buffer.String(), "\r\n") {
				return buffer.String(), nil
			}
		}
	}
	return buffer.String(), nil
}

func (s *Socket) run(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		select {
		case <-s.shutdown:
			s.wg.Done()
			return
		default:
			msg, err := reader.ReadString('\n')
			if err != nil {
				s.Close()
				log.Fatalf("incoming message error: %v\n", err)
			}
			s.incoming <- msg
		}
	}
}
