package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type Vote struct {
	Header    VoteHeader
	Signature []byte
}
type VoteHeader struct {
	From        []byte
	Epoch       int
	Block_num   int
	Shard_level int
	Ifagree     bool
}

func NewVote(from []byte, epoch int, block_num int, shard_level int, ifagree bool) *Vote {

	v := Vote{Header: VoteHeader{From: from, Epoch: epoch, Block_num: block_num, Shard_level: shard_level, Ifagree: ifagree}}

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

func (v *Vote) VerifyVote(pow []byte) bool {

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
