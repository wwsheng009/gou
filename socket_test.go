package gou

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadSocket(t *testing.T) {
	sock, err := LoadSocket("file://"+path.Join(TestServerRoot, "rfid.sock.json"), "rfid")
	assert.Nil(t, err)
	assert.Equal(t, sock.Name, "RFID Receiver (Server Mode)")
	assert.Equal(t, sock.Mode, "server")
	assert.Equal(t, sock.Event.Data, "scripts.socket.onData")
	assert.Equal(t, sock.Event.Closed, "scripts.socket.onClosed")
	assert.Equal(t, sock.Event.Error, "scripts.socket.onError")
	assert.Equal(t, sock.Event.Connected, "scripts.socket.onConnected")
	assert.Equal(t, sock.Port, "3019")
	assert.Equal(t, sock.Host, "0.0.0.0")
}

func TestSocketStart(t *testing.T) {
	LoadSocket("file://"+path.Join(TestServerRoot, "rfid.sock.json"), "rfid")
	sock := SelectSocket("rfid")
	assert.Equal(t, sock.Name, "RFID Receiver (Server Mode)")
	assert.Equal(t, sock.Mode, "server")
	assert.Equal(t, sock.Event.Data, "scripts.socket.onData")
	assert.Equal(t, sock.Event.Closed, "scripts.socket.onClosed")
	assert.Equal(t, sock.Event.Error, "scripts.socket.onError")
	assert.Equal(t, sock.Event.Connected, "scripts.socket.onConnected")
	assert.Equal(t, sock.Port, "3019")
	assert.Equal(t, sock.Host, "0.0.0.0")

	// sock.Start()
}

func TestSocketConnect(t *testing.T) {
	LoadSocket("file://"+path.Join(TestServerRoot, "rfid_client.sock.json"), "rfid_client")
	sock := SelectSocket("rfid_client")
	assert.Equal(t, sock.Name, "RFID Receiver (Client Mode)")
	assert.Equal(t, sock.Mode, "client")
	assert.Equal(t, sock.Event.Data, "scripts.socket.onData")
	assert.Equal(t, sock.Event.Error, "scripts.socket.onError")
	assert.Equal(t, sock.Event.Closed, "scripts.socket.onClosed")
	assert.Equal(t, sock.Event.Connected, "scripts.socket.onConnected")
	assert.Equal(t, sock.Port, "3019")
	assert.Equal(t, sock.Host, "192.168.31.33")

	// sock.Connect()
}

func TestSocketOpen(t *testing.T) {
	// err := Yao.Load(path.Join(TestScriptRoot, "socket.js"), "socket")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// LoadSocket("file://"+path.Join(TestServerRoot, "rfid_client.sock.json"), "rfid_client")
	// sock := SelectSocket("rfid_client")
	// sock.Open()
}
