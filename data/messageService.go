package data

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/redsuperbat/nano-flow/logging"
	"go.uber.org/zap"
)

func InitDatabase(filepath string) (*os.File, error) {
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
		return nil, fmt.Errorf("error checking database existence: %s", err)
	} else {
		logger.Infof("database '%s' already exists", filepath)
	}
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

type DbChannel = chan uint8

func InitWatcher(filepath string) (DbChannel, error) {
	logger := logging.New()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	channel := make(DbChannel)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					channel <- 0
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Errorln(err)
			}
		}
	}()

	err = watcher.Add(filepath)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func NewMessageService(file *os.File) *MessageService {
	logger := logging.New()
	return &MessageService{
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

func (s *MessageService) Subscribe(func(*Message)) {

}
