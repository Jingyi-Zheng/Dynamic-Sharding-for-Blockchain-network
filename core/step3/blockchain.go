package main

import (
	"fmt"
	"time"
)

type TransactionsQueue chan *Transaction
type BlocksQueue chan Block
type VoteQueue chan Vote
type StateQueue chan State

type Blockchain struct {
	CurrentBlock Block
	BlockSlice

	TransactionsQueue
	BlocksQueue
	VoteQueue
	StateQueue
}

const (
	MESSAGE_GET_NODES = iota + 128
	MESSAGE_SEND_NODES

	MESSAGE_GET_TRANSACTION
	MESSAGE_SEND_TRANSACTION

	MESSAGE_GET_BLOCK
	MESSAGE_SEND_BLOCK

	MESSAGE_GET_VOTE
	MESSAGE_SEND_VOTE

	MESSAGE_GET_STATE
	MESSAGE_SEND_STATE
)

const (
	Nodenum         = 6
	Firstshardsize  = 3
	Secondshardsize = 6
	t1              = 200
	t2              = 2000
)

func SetupBlockchain() *Blockchain {

	bl := new(Blockchain)
	bl.TransactionsQueue, bl.BlocksQueue = make(TransactionsQueue), make(BlocksQueue)

	//Read blockchain from file and stuff...

	bl.CurrentBlock = bl.CreateNewBlock()

	return bl
}

func (bl *Blockchain) CreateNewBlock() Block {

	prevBlock := bl.BlockSlice.PreviousBlock()
	prevBlockHash := []byte{}
	if prevBlock != nil {
		prevBlockHash = prevBlock.Hash()
	}

	b := NewBlock(prevBlockHash)
	b.BlockHeader.From = Core.Keypair.Public

	return b
}

func (bl *Blockchain) AddBlock(b Block) {

	bl.BlockSlice = append(bl.BlockSlice, b)
}

func (bl *Blockchain) Run(mynetwork Network) {

	interruptBlockGen := bl.GenerateBlocks()
	var nowepoch *Epoch
	nowepochint := 0
	go func() {
		for {
			nowepoch = NewEpoch(uint32(nowepochint))
			time.Sleep(t1 * time.Second)
			nowepoch.Intimebound1 = false
			time.Sleep(t2 * time.Second)
			nowepoch.Intimebound2 = false
			nowepochint = nowepochint + 1
		}
	}()
	for {
		select {
		case tr := <-bl.TransactionsQueue:

			if bl.CurrentBlock.TransactionSlice.Exists(*tr) {
				continue
			}
			if !tr.VerifyTransaction() {
				fmt.Println("Recieved non valid transaction", tr)
				continue
			}

			bl.CurrentBlock.AddTransaction(tr)
			interruptBlockGen <- bl.CurrentBlock

			//Broadcast transaction to the network
			mes := NewMessage(MESSAGE_SEND_TRANSACTION)
			mes.Data, _ = tr.Transaction2bytes()

			time.Sleep(300 * time.Millisecond)
			mynetwork.BroadcastQueue_level1 <- *mes

		case b := <-bl.BlocksQueue:
			if b.Epoch != nowepoch.Epoch {
				continue
			}
			if nowepoch.Block.Exists(b) || nowepoch.PendingBlock.Exists(b) {
				fmt.Println("block exists")
				continue
			}
			if !b.VerifyBlock() {
				fmt.Println("block verification fails")
				v := NewVote(Core.Keypair.Public, nowepoch.Epoch, b.Signature, false)
				v.Signature = v.Signvote(Core.Keypair)
				mes := NewMessage(MESSAGE_SEND_VOTE)
				mes.Data, _ = v.Vote2bytes()
				mynetwork.BroadcastQueue_level2 <- *mes
				continue
			}

			fmt.Println("New block in pending!", b.Hash())
			nowepoch.AddPendingBlock(b)

			//Broadcast block
			mes := NewMessage(MESSAGE_SEND_BLOCK)
			mes.Data, _ = b.Block2bytes()
			if nowepoch.Intimebound1 == true {
				Core.Network.BroadcastQueue_level1 <- *mes
			} else {
				Core.Network.BroadcastQueue_level2 <- *mes
			}

			v := NewVote(Core.Keypair.Public, nowepoch.Epoch, b.Signature, true)
			mes2 := NewMessage(MESSAGE_SEND_VOTE)
			v.Signature = v.Signvote(Core.Keypair)
			mes2.Data, _ = v.Vote2bytes()
			mynetwork.BroadcastQueue_level2 <- *mes2
			//New Block
			// bl.CurrentBlock = bl.CreateNewBlock()
			// bl.CurrentBlock.TransactionSlice = &transDiff

			// interruptBlockGen <- bl.CurrentBlock

		case vote := <-bl.VoteQueue:
			if vote.VerifyVote() == false || vote.Header.Epoch != nowepoch.Epoch {
				continue
			}
			voted_block := nowepoch.PendingBlock.Find(vote.Header.Block_sign)
			if (voted_block != nil) && voted_block.Votepool.Exists(vote) == false {
				voted_block.AddVote(&vote)
				mes := NewMessage(MESSAGE_SEND_VOTE)
				mes.Data, _ = vote.Vote2bytes()
				mynetwork.BroadcastQueue_level2 <- *mes
			}
			if voted_block == nil {
				bl.VoteQueue <- vote
				continue
			}
			if (len(voted_block.Votepool) == Firstshardsize) && nowepoch.Intimebound1 == true {
				fmt.Println("New block !", voted_block.Hash())
				nowepoch.AddBlock(*voted_block)
				//todo
				//Read the state of last epoch and tx in this block, then create new state
				st := NewState(nowepoch.Epoch)
				mes := NewMessage(MESSAGE_SEND_STATE)
				mes.Data, _ = st.State2bytes()
				mynetwork.BroadcastQueue_level2 <- *mes
				continue
			}
			if (len(voted_block.Votepool) > Secondshardsize/2) && nowepoch.Intimebound2 == true {
				fmt.Println("New block !", voted_block.Hash())
				nowepoch.AddBlock(*voted_block)
				continue
			}

		case st := <-bl.StateQueue:

		}

	}
}

func (bl *Blockchain) GenerateBlocks() chan Block {

	interrupt := make(chan Block)

	go func() {

		block := <-interrupt
	loop:
		fmt.Println("Starting Proof of Work...")
		block.BlockHeader.MerkelRoot = block.GenerateMerkelRoot()
		block.BlockHeader.Nonce = 0
		block.BlockHeader.Timestamp = uint32(time.Now().Unix())

		for true {

			sleepTime := time.Nanosecond
			if block.TransactionSlice.Len() > 0 {

				if CheckProofOfWork(BLOCK_POW, block.Hash()) {

					block.Signature = block.Sign(Core.Keypair)
					bl.BlocksQueue <- block
					sleepTime = time.Hour * 24
					fmt.Println("Found Block!")

				} else {

					block.BlockHeader.Nonce += 1
				}

			} else {
				sleepTime = time.Hour * 24
				fmt.Println("No trans sleep")
			}

			select {
			case block = <-interrupt:
				goto loop
			case <-Timeout(sleepTime):
				continue
			}
		}
	}()

	return interrupt
}
