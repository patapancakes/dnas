package main

import (
	"encoding/binary"
	"io"
)

type messageHeader struct {
	Flags    uint16
	Opcode   uint16
	Security [32]byte
	DataLen  uint32
}

type message struct {
	messageHeader
	Data []byte
}

func readMessage(r io.Reader) (message, error) {
	var m message
	err := binary.Read(r, binary.BigEndian, &m.messageHeader)
	if err != nil {
		return message{}, err
	}

	m.Data = make([]byte, m.DataLen)
	_, err = io.ReadFull(r, m.Data)
	return m, err
}

func (m message) writeTo(w io.Writer) (int64, error) {
	m.DataLen = uint32(len(m.Data))

	err := binary.Write(w, binary.BigEndian, m.messageHeader)
	if err != nil {
		return 0, err
	}

	n, err := w.Write(m.Data)
	return int64(binary.Size(m.messageHeader) + n), err
}
