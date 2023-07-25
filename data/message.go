package data

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"time"

	"github.com/redsuperbat/nano-flow/logging"
	"go.uber.org/zap"
)

const (
	HEADER_LENGTH = 18
)

func Init(filepath string) (*os.File, error) {
	_, err := os.Stat(filepath)
	logger := logging.New()
	if os.IsNotExist(err) {
		file, err := os.Create(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %s", err)
		}
		logger.Infof("database '%s' created successfully", filepath)
		return file, nil
	} else if err != nil {
		return nil, fmt.Errorf("error checking file existence: %s", err)
	} else {
		logger.Infof("database file '%s' already exists", filepath)
	}
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

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

func NewMessageService(file *os.File) MessageService {
	logger := logging.New()
	return MessageService{
		DatabaseHandle: file,
		logger:         logger,
	}
}

type MessageService struct {
	DatabaseHandle *os.File
	logger         *zap.SugaredLogger
}

func (ms *MessageService) AppendMessage(message *Message) (*Message, error) {
	fh := ms.DatabaseHandle
	_, err := fh.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, err
	}
	data := message.marshal()
	_, err = fh.Write(data)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (ms *MessageService) GetAllMessages() ([]Message, error) {
	ms.DatabaseHandle.Seek(0, 0)
	buf, err := io.ReadAll(ms.DatabaseHandle)
	if err != nil {
		return nil, err
	}
	messages := []Message{}
	i := 0
	for {
		if i >= len(buf) {
			break
		}
		startIndex := i + 2
		endIndex := i + 6
		clBuf := buf[startIndex:endIndex]
		contentLength := binary.BigEndian.Uint32(clBuf)
		endIndex = i + int(contentLength) + HEADER_LENGTH
		messageData := buf[i:endIndex]
		i = endIndex
		message, err := ParseMessage(messageData)
		if err != nil {
			ms.logger.Panicln("corrupted message data")
		}
		messages = append(messages, message)
	}
	return messages, nil

}
