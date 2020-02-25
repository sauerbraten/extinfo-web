package cubecode

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrBufferTooShort = errors.New("cubecode: buffer too short")

// Packet represents a Sauerbraten UDP packet.
type Packet struct {
	buf *bytes.Buffer
}

// NewPacket returns a Packet using buf as the underlying buffer.
func NewPacket(buf []byte) *Packet {
	return &Packet{
		buf: bytes.NewBuffer(buf),
	}
}

// Len returns the number of unread bytes left in the packet.
func (p *Packet) Len() int {
	return p.buf.Len()
}

// HasRemaining returns true if there are bytes remaining to be read in the packet.
func (p *Packet) HasRemaining() bool {
	return p.Len() > 0
}

// SubPacketFromRemaining returns a packet from the bytes remaining in p's buffer.
func (p *Packet) SubPacketFromRemaining() (*Packet, error) {
	return p.SubPacket(p.Len())
}

// SubPacket returns a part of the packet as new packet. It does not copy the underlying slice!
func (p *Packet) SubPacket(n int) (q *Packet, err error) {
	if n > p.Len() {
		return nil, fmt.Errorf("cubecode: sub-packet of length %v requested, but there are only %v bytes left", n, p.Len())
	}

	return NewPacket(p.buf.Next(n)), nil
}

func (p *Packet) WriteByte(b byte) error {
	err := p.buf.WriteByte(b)
	if err != nil {
		err = fmt.Errorf("cubecode: %s", err)
	}
	return err
}

// WriteInt writes an int32 to the packet buffer. It grows the packet as needed.
func (p *Packet) WriteInt(value int32) {
	if value < 128 && value > -127 {
		p.buf.WriteByte(byte(value))
	} else if value < 0x8000 && value >= -0x8000 { // TODO: check this
		p.buf.WriteByte(0x80)
		p.buf.WriteByte(byte(value))
		p.buf.WriteByte(byte(value >> 8))
	} else {
		p.buf.WriteByte(0x81)
		p.buf.WriteByte(byte(value))
		p.buf.WriteByte(byte(value >> 8))
		p.buf.WriteByte(byte(value >> 16))
		p.buf.WriteByte(byte(value >> 24))
	}
}

func (p *Packet) WriteString(s string) {
	for _, c := range s {
		p.WriteInt(uni2Cube[c])
	}
	p.WriteByte(0x00)
}

// ReadByte returns the next byte in the packet.
func (p *Packet) ReadByte() (byte, error) {
	b, err := p.buf.ReadByte()
	if err != nil {
		err = fmt.Errorf("cubecode: %s", err)
	}
	return b, err
}

// ReadInt returns the integer value encoded in the next bytes of the packet.
func (p *Packet) ReadInt() (int, error) {
	n := p.Len()

	if n < 1 {
		return -1, ErrBufferTooShort
	}

	var (
		value int32
		err   error
		b     byte
	)

	b, err = p.ReadByte()
	if err != nil {
		return -1, err
	}

	switch b {
	default:
		// most often, th value is only one byte
		// convert to int and keep the sign of the 8-bit representation
		value = int32(int8(b))

	case 0x80:
		// value is contained in the next two bytes
		value, err = p.readInt(2)
		// make sure to keep the sign of the 16-bit representation
		value = int32(int16(value))

	case 0x81:
		// value is contained in the next four bytes
		value, err = p.readInt(4)

	}

	return int(value), err
}

// readInt reads n bufNewbytes from the packet and decodes them into an int value.
func (p *Packet) readInt(n int) (int32, error) {
	var (
		value int32
		err   error
		b     byte
	)

	for i := 0; i < n; i++ {
		b, err = p.ReadByte()
		if err != nil {
			return -1, err
		}

		value += int32(b) << (8 * uint(i))
	}

	return value, nil
}

// ReadString returns a string of the next bytes up to 0x00.
func (p *Packet) ReadString() (s string, err error) {
	var value int
	value, err = p.ReadInt()
	if err != nil {
		return
	}

	for value != 0x00 {
		codepoint := uint8(value)

		s += string(cubeToUni[codepoint])

		value, err = p.ReadInt()
		if err != nil {
			return
		}
	}

	return
}

// Matches sauer color codes (sauer uses form feed followed by a digit, e.g. \f3 for red)
var sauerStringsSanitizer = regexp.MustCompile("\\f.")

// SanitizeString returns the string, cleared of sauer color codes like \f3 for red.
func SanitizeString(s string) string {
	s = sauerStringsSanitizer.ReplaceAllLiteralString(s, "")
	return strings.TrimSpace(s)
}
