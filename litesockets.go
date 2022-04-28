package litesockets

import (
	"encoding/binary"
	"net"
	"time"
)

type Socket struct {
	conn    net.Conn
	timeout time.Duration
}

func OpenSocket(address string, timeout time.Duration) (*Socket, error) {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, err
	}
	return &Socket{
		conn:    conn,
		timeout: timeout,
	}, nil
}

func (s *Socket) Read() ([]byte, error) {
	s.conn.SetReadDeadline(time.Now().Add(s.timeout))
	lengthBytes := make([]byte, 8)
	var msgLen uint64
	var index int
	for index < 8 {
		n, err := s.conn.Read(lengthBytes[index:])
		if err != nil {
			return nil, err
		}
		index += n
	}
	msgLen = binary.LittleEndian.Uint64(lengthBytes)
	msg := make([]byte, msgLen)
	index = 0
	for index < len(msg) {
		n, err := s.conn.Read(msg[index:])
		if err != nil {
			return nil, err
		}
		index += n
	}
	return msg, nil
}

func (s *Socket) Write(msg []byte) (int, error) {
	if msg == nil {
		s.conn.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
		return 0, nil
	}
	s.conn.SetWriteDeadline(time.Now().Add(s.timeout))
	lengthBytes := make([]byte, 8)
	msgLen := uint64(len(msg))
	binary.LittleEndian.PutUint64(lengthBytes, msgLen)
	var written int
	n, err := s.conn.Write(lengthBytes)
	written += n
	if err != nil {
		return written, err
	}
	n, err = s.conn.Write(msg)
	written += n

	return written, err
}

func (s *Socket) Close() error {
	return s.conn.Close()
}

func (s *Socket) LocalAddress() net.Addr {
	return s.conn.LocalAddr()
}

func (s *Socket) RemoteAddress() net.Addr {
	return s.conn.RemoteAddr()
}

type SocketServer interface {
	BeginServing()
}

type SimpleSocketServer struct {
	listener net.Listener
	handler  func(socket *Socket)
	timeout  time.Duration
	Errors   chan error
}

func NewSimpleSocketServer(address string, errorChannelSize int, timeout time.Duration, handler func(socket *Socket)) (*SimpleSocketServer, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &SimpleSocketServer{
		listener: listener,
		handler:  handler,
		timeout:  timeout,
		Errors:   make(chan error, errorChannelSize),
	}, nil
}

func (s *SimpleSocketServer) BeginServing() {
	defer s.listener.Close()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case s.Errors <- err:
			default:
			}
		} else {
			go s.handler(&Socket{conn: conn, timeout: s.timeout})
		}
	}
}
