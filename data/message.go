package data

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"time"
)

const (
	HEADER_LENGTH = 18
)

type MessageChannel = chan *Message

type Message struct {
	// Header
	Version       uint16
	ContentLength uint32
	Timestamp     int64
	Crc           uint32
	// Body
	Data []byte
}

func (m *Message) marshal() []byte {
	version := make([]byte, 2)
	binary.BigEndian.PutUint16(version, m.Version)
	cl := make([]byte, 4)
	binary.BigEndian.PutUint32(cl, m.ContentLength)
	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, uint64(m.Timestamp))
	crc := make([]byte, 4)
	binary.BigEndian.PutUint32(crc, m.Crc)
	x := append(version, cl...)
	x = append(x, timestamp...)
	x = append(x, crc...)
	x = append(x, m.Data...)
	return x
}

func NewMessage(data []byte) Message {
	var v uint16 = 0
	version := make([]byte, 2)
	binary.BigEndian.PutUint16(version, v)
	contentLength := uint32(len(data))

	ts := time.Now().UTC().UnixNano()
	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, uint64(ts))
	cl := make([]byte, 4)
	binary.BigEndian.PutUint32(cl, contentLength)
	bytes := append([]byte{}, version...)
	bytes = append(bytes, cl...)
	bytes = append(bytes, timestamp...)
	bytes = append(bytes, data...)

	crc32Hash := crc32.NewIEEE()
	crc32Hash.Write(bytes)
	crc := crc32Hash.Sum32()

	return Message{
		Version:       v,
		ContentLength: contentLength,
		Timestamp:     ts,
		Crc:           crc,
		Data:          data,
	}
}

func ParseMessage(content []byte) (Message, error) {
	var msg Message
	msg.Version = binary.BigEndian.Uint16(content[0:2])
	msg.ContentLength = binary.BigEndian.Uint32(content[2:6])
	if len(content) != HEADER_LENGTH+int(msg.ContentLength) {
		return msg, errors.New("invalid message content")
	}
	msg.Timestamp = int64(binary.BigEndian.Uint64(content[6:14]))
	msg.Crc = binary.BigEndian.Uint32(content[14:HEADER_LENGTH])
	msg.Data = content[HEADER_LENGTH:(HEADER_LENGTH + int(msg.ContentLength))]
	return msg, nil
}
