package data

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	// Do nothing for now
}

func teardown() {
	os.Remove("test.db")
}

func TestCreateData(t *testing.T) {
	file, _ := InitDatabase("test.db")
	service := NewMessageService(file)
	msg1 := NewMessage([]byte("Hi there my old friend!"))
	msg2 := NewMessage([]byte("You're a wizard harry!"))
	service.AppendMessage(&msg1)
	service.AppendMessage(&msg2)
	messages, _ := service.GetAllMessages()

	if len(messages) != 2 {
		t.Errorf("expected messages to have been created. Len should equal '2' got %d", len(messages))
	}

}
