package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
)

type BlockSlice []Block

func (bs BlockSlice) Exists(b Block) bool {

	//Traverse array in reverse order because if a block exists is more likely to be on top.
	l := len(bs)
	for i := l - 1; i >= 0; i-- {

		bb := bs[i]
		if reflect.DeepEqual(b.Signature, bb.Signature) {
			return true
		}
	}

	return false
}

func (bs BlockSlice) Find(Sign []byte) *Block {
	l := len(bs)
	for i := l - 1; i >= 0; i-- {

		bb := bs[i]
		if reflect.DeepEqual(Sign, bb.Signature) {
			return &bb
		}
	}
	return nil
}

func (bs BlockSlice) PreviousBlock() *Block {
	l := len(bs)
	if l == 0 {
		return nil
	} else {
		return &bs[l-1]
	}
}

type Votepool []Vote
type Block struct {
	*BlockHeader
	Signature []byte
	*TransactionSlice
	Votepool Votepool
}

type BlockHeader struct {
	Epoch       uint32
	Index       uint32
	Shard_level uint32
	From        []byte
	PrevBlock   []byte
	MerkelRoot  []byte
	Timestamp   uint32
}

func NewBlock(previousBlock []byte) Block {

	header := &BlockHeader{PrevBlock: previousBlock}
	return Block{header, nil, new(TransactionSlice), nil}
}

func (b *Block) AddVote(v *Vote) {
	b.Votepool = append(b.Votepool, *v)
}

func (b *Block) AddTransaction(t *Transaction) {
	newSlice := b.TransactionSlice.AddTransaction(*t)
	b.TransactionSlice = &newSlice
}

func (b *Block) Sign(keypair *Keypair) []byte {

	s, _ := keypair.Sign(b.Hash())
	return s
}

func (b *Block) VerifyBlock() bool {

	headerHash := b.Hash()
	merkel := b.GenerateMerkelRoot()

	return reflect.DeepEqual(merkel, b.BlockHeader.MerkelRoot) && SignatureVerify(b.BlockHeader.From, b.Signature, headerHash)
}

func (b *Block) Hash() []byte {

	headerHash, _ := b.BlockHeader.Blockheader2bytes()
	return SHA256(headerHash)
}

func (b *Block) GenerateMerkelRoot() []byte {

	var merkell func(hashes [][]byte) []byte
	merkell = func(hashes [][]byte) []byte {

		l := len(hashes)
		if l == 0 {
			return nil
		}
		if l == 1 {
			return hashes[0]
		} else {

			if l%2 == 1 {
				return merkell([][]byte{merkell(hashes[:l-1]), hashes[l-1]})
			}

			bs := make([][]byte, l/2)
			for i, _ := range bs {
				j, k := i*2, (i*2)+1
				bs[i] = SHA256(append(hashes[j], hashes[k]...))
			}
			return merkell(bs)
		}
	}

	ts := Map(func(t Transaction) []byte { return t.Hash() }, []Transaction(*b.TransactionSlice)).([][]byte)
	return merkell(ts)

}
func (b *Block) Block2bytes() ([]byte, error) {

	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(b)
	if err != nil {
		fmt.Println(err)

	}
	return buf.Bytes(), nil

}

func (b *Block) Bytes2block(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	decoder.Decode(&b)
	return nil
}

func (h *BlockHeader) Blockheader2bytes() ([]byte, error) {

	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(h)
	if err != nil {
		fmt.Println(err)

	}
	return buf.Bytes(), nil
}

func (h *BlockHeader) Bytes2blockheader(data []byte) error {

	decoder := gob.NewDecoder(bytes.NewReader(data))
	decoder.Decode(&h)
	return nil

	return nil
}
