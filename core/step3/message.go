package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type Message struct {
	// transaction? block? state? vote?
	Identifier byte
	Options    []byte
	Data       []byte
	Reply      chan Message
}

func NewMessage(id byte) *Message {

	return &Message{Identifier: id}
}

func (m *Message) Message2Bytes() ([]byte, error) {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(m)
	if err != nil {
		fmt.Println(err)

	}
	return (buf.Bytes()), nil
}

func (m *Message) Bytes2Message(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	decoder.Decode(m)

	return nil
}
