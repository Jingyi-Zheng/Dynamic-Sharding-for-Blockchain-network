package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"time"
)

type Transaction struct {
	Header    TransactionHeader
	Signature []byte
	Payload   []byte
}

type TransactionHeader struct {
	From        []byte
	To          []byte
	Timestamp   uint32
	PayloadHash []byte
}

// Returns bytes to be sent to the network
func NewTransaction(from, to, payload []byte) *Transaction {

	t := Transaction{Header: TransactionHeader{From: from, To: to}, Payload: payload}

	t.Header.Timestamp = uint32(time.Now().Unix())
	t.Header.PayloadHash = SHA256(t.Payload)

	return &t
}

func (t *Transaction) Hash() []byte {

	headerBytes, _ := t.Header.TransactionHeader2bytes()
	return SHA256(headerBytes)
}

func (t *Transaction) Sign(keypair *Keypair) []byte {

	s, _ := keypair.Sign(t.Hash())

	return s
}

func (t *Transaction) VerifyTransaction() bool {

	headerHash := t.Hash()
	payloadHash := SHA256(t.Payload)

	return reflect.DeepEqual(payloadHash, t.Header.PayloadHash) && SignatureVerify(t.Header.From, t.Signature, headerHash)
}

func (t *Transaction) Transaction2bytes() ([]byte, error) {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(t)
	if err != nil {
		fmt.Println(err)

	}
	return buf.Bytes(), err

}

func (t *Transaction) Bytes2Transaction(data []byte) error {

	decoder := gob.NewDecoder(bytes.NewReader(data))
	decoder.Decode(&t)
	return nil

}

func (th *TransactionHeader) TransactionHeader2bytes() ([]byte, error) {

	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(th)
	if err != nil {
		fmt.Println(err)

	}
	return buf.Bytes(), err

}

func (th *TransactionHeader) Bytes2Transactionheader(data []byte) error {

	decoder := gob.NewDecoder(bytes.NewReader(data))
	decoder.Decode(&th)
	return nil

}

type TransactionSlice []Transaction

func (slice TransactionSlice) Len() int {

	return len(slice)
}

func (slice TransactionSlice) Exists(tr Transaction) bool {

	for _, t := range slice {
		if reflect.DeepEqual(t.Signature, tr.Signature) {
			return true
		}
	}
	return false
}

func (slice TransactionSlice) AddTransaction(t Transaction) TransactionSlice {

	// Inserted sorted by timestamp
	for i, tr := range slice {
		if tr.Header.Timestamp >= t.Header.Timestamp {
			return append(append(slice[:i], t), slice[i:]...)
		}
	}

	return append(slice, t)
}

func (slice *TransactionSlice) TransactionSlice2bytes() ([]byte, error) {

	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(slice)
	if err != nil {
		fmt.Println(err)

	}
	return buf.Bytes(), err
}

func (slice *TransactionSlice) Bytes2TransactionSlice(data []byte) error {

	decoder := gob.NewDecoder(bytes.NewReader(data))
	decoder.Decode(&slice)
	return nil
}
