package data

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/redsuperbat/nano-flow/logging"
	"go.uber.org/zap"
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
	Len     uint32
	Data    []byte
	Version uint8
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

func (ms *MessageService) AppendMessage(data []byte) error {
	l := uint32(len(data))
	fh := ms.DatabaseHandle
	_, err := fh.Seek(0, os.SEEK_END)
	if err != nil {
		return err
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, l)
	_, err = fh.Write(append(buf, data...))
	if err != nil {
		return err
	}

	return nil
}

func (ms *MessageService) GetAllMessages() ([]Message, error) {
	buf, err := io.ReadAll(ms.DatabaseHandle)
	if err != nil {
		return nil, err
	}

	len := binary.LittleEndian.Uint32(buf[:4])
	ms.logger.Infof("len %d", len)
	return []Message{}, nil

}

func (ms *MessageService) PrintAllMessages() {

	ms.DatabaseHandle.Seek(0, 0)
	bufferSize := 1024
	buffer := make([]byte, bufferSize)

	for {
		n, err := ms.DatabaseHandle.Read(buffer)
		if err != nil {
			ms.logger.Infoln(err)
			break
		}
		ms.logger.Infoln(string(buffer[:n]))
	}
}
