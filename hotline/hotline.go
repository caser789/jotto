package hotline

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

// Hotline provides simpler interface to interact with
// TCP connection communicating via Protobuf messages.
// Boundaries of Protobuf messages are identified by an
// integer of the message size sent before each message.
type Hotline struct {
	conn    net.Conn
	alive   bool
	timeout time.Duration
}

// NewHotline creates a new hotline instance.
func NewHotline(conn net.Conn, timeout time.Duration) *Hotline {
	return &Hotline{
		conn:    conn,
		alive:   true,
		timeout: timeout,
	}
}

func (p *Hotline) String() string {
	return fmt.Sprintf("(remote:%s,timeout:%s)", p.conn.RemoteAddr(), p.timeout)
}

// Write() writes the given message to the underlying TCP connection.
func (p *Hotline) Write(kind uint32, message []byte) (err error) {
	header := make([]byte, 8)
	size := len(message)

	if size <= 0 {
		return errors.New("Invalid message size")
	}

	binary.LittleEndian.PutUint32(header, uint32(kind))
	binary.LittleEndian.PutUint32(header[4:], uint32(size))

	buffer := make([]byte, 8+size)

	// Two copies? Or two syscalls?
	copy(buffer, header)
	copy(buffer[8:], message)

	p.conn.SetWriteDeadline(time.Now().Add(p.timeout))
	_, err = p.conn.Write(buffer)

	if err != nil {
		p.alive = false
		return
	}

	return
}

// Read() reads a message from the underlying TCP connection.
func (p *Hotline) Read() (uint32, []byte, error) {
	header := make([]byte, 8)

	p.conn.SetReadDeadline(time.Now().Add(p.timeout))
	_, err := p.conn.Read(header)

	if err != nil {
		p.alive = false
		return 0, nil, err
	}

	kind := binary.LittleEndian.Uint32(header)
	size := binary.LittleEndian.Uint32(header[4:])

	if size <= 0 {
		return 0, nil, errors.New("Invalid message size")
	}

	buffer := make([]byte, size)
	p.conn.SetReadDeadline(time.Now().Add(p.timeout))
	_, err = p.conn.Read(buffer)

	if err != nil {
		p.alive = false
		return 0, nil, err
	}

	return kind, buffer, nil
}

func (p *Hotline) Close() {
	p.conn.Close()
	p.alive = false
}
