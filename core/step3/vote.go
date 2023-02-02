package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
)

type Vote struct {
	Header    VoteHeader
	Signature []byte
}

type VoteHeader struct {
	From       []byte
	Epoch      uint32
	Block_sign []byte
	Ifagree    bool
}

func NewVote(from []byte, epoch uint32, block_sign []byte, ifagree bool) *Vote {

	v := Vote{Header: VoteHeader{From: from, Epoch: epoch, Block_sign: block_sign, Ifagree: ifagree}}

	return &v
}

func (v *Vote) Hash() []byte {

	headerBytes, _ := v.Header.VoteHeader2bytes()
	return SHA256(headerBytes)
}

func (v *Vote) Signvote(keypair *Keypair) []byte {

	s, _ := keypair.Sign(v.Hash())

	return s
}

func (vp Votepool) Exists(vote Vote) bool {

	for _, v := range vp {
		if reflect.DeepEqual(v.Signature, vote.Signature) {
			return true
		}
	}
	return false
}

func (v *Vote) VerifyVote() bool {

	headerHash := v.Hash()

	return SignatureVerify(v.Header.From, v.Signature, headerHash)
}

func (v *Vote) Vote2bytes() ([]byte, error) {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(v)
	if err != nil {
		fmt.Println(err)

	}
	return buf.Bytes(), err

}

func (v *Vote) Bytes2Vote(data []byte) error {

	decoder := gob.NewDecoder(bytes.NewReader(data))
	decoder.Decode(&v)
	return nil

}

func (vh *VoteHeader) VoteHeader2bytes() ([]byte, error) {

	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(vh)
	if err != nil {
		fmt.Println(err)

	}
	return buf.Bytes(), err

}

func (vh *VoteHeader) Bytes2Voteheader(data []byte) error {

	decoder := gob.NewDecoder(bytes.NewReader(data))
	decoder.Decode(&vh)
	return nil

}
