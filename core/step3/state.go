package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type State struct {
	Header    StateHeader
	Signature []byte
}

type StateHeader struct {
	From     []byte
	Epoch    uint32
	Statemap map[string]float64
	Stateid  []byte
}

func NewState(Epoch uint32) *State {
	return &State{Header: StateHeader{nil, Epoch, nil, nil}}
}

func (st *State) SetStateid() {
	mapbytes, _ := st.Statemap2bytes()
	st.Header.Stateid = SHA256(mapbytes)
}

func (St State) MergeState(St2 *State) {
	for k, v := range *&St2.Header.Statemap {
		St.Header.Statemap[k] = v
	}
}

func (st *State) Hash() []byte {

	headerBytes, _ := st.Header.Stateheader2bytes()
	return SHA256(headerBytes)
}

func (st *State) VerifyState() bool {

	headerHash := st.Hash()

	return SignatureVerify(st.Header.From, st.Signature, headerHash)
}

func (st *State) Signvote(keypair *Keypair) []byte {

	s, _ := keypair.Sign(st.Hash())

	return s
}
func (st *State) State2bytes() ([]byte, error) {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(st)
	if err != nil {
		fmt.Println(err)

	}
	return buf.Bytes(), err

}
func (st *State) Statemap2bytes() ([]byte, error) {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(st.Header.Statemap)
	if err != nil {
		fmt.Println(err)

	}
	return buf.Bytes(), err

}
func (st *StateHeader) Stateheader2bytes() ([]byte, error) {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(st)
	if err != nil {
		fmt.Println(err)

	}
	return buf.Bytes(), err

}
func (st *State) Bytes2State(data []byte) error {

	decoder := gob.NewDecoder(bytes.NewReader(data))
	decoder.Decode(&st)
	return nil

}
