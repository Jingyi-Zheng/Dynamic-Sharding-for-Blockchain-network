package main

type Epoch struct {
	Epoch        uint32
	Block        BlockSlice
	State        Statemap
	PendingBlock BlockSlice
	RecState     StateSlice
	Intimebound1 bool
	Intimebound2 bool
}

func NewEpoch(epoch uint32) *Epoch {
	return &Epoch{epoch, nil, nil, nil, nil, true, true}
}

func (ep *Epoch) AddBlock(block Block) {
	ep.Block = append(ep.Block, block)
}

func (ep *Epoch) AddPendingBlock(block Block) {
	ep.PendingBlock = append(ep.PendingBlock, block)
}

func (ep *Epoch) AddRecState(st State) {
	ep.RecState = append(ep.RecState, st)
}
