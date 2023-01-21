package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"math/rand"
	"os"
)

const Nodenum = 6
const firstshardsize = 3
const secondshardsize = 6

type nodelist struct {
	Nodes  [Nodenum]string
	Random string
}

type nodelist_shard struct {
	Nodes              [Nodenum]string
	FirstShard         [Nodenum]int
	SecondShard        [Nodenum]int
	FirstShard_leader  [Nodenum]int
	SecondShard_leader [Nodenum]int
}

func hash(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())
}

func main() {
	tempbyte, err := os.ReadFile("../step1/nodes/nodes&RN.gob")
	if err != nil {
		log.Fatal(err)
	}
	var step1db nodelist
	decoder := gob.NewDecoder(bytes.NewReader(tempbyte))
	decoder.Decode(&step1db)
	fmt.Println(step1db)
	rand.Seed((hash(step1db.Random)))
	fmt.Println(step1db.Nodes)
	rand.Shuffle(len(step1db.Nodes), func(i, j int) { step1db.Nodes[i], step1db.Nodes[j] = step1db.Nodes[j], step1db.Nodes[i] })
	var firstShard [Nodenum]int
	var secondShard [Nodenum]int
	var firstShard_leader [Nodenum]int
	var secondShard_leader [Nodenum]int
	for i := 0; i < len(step1db.Nodes); i++ {
		firstShard[i] = int(math.Ceil(float64(i+1) / float64(firstshardsize)))
		secondShard[i] = int(math.Ceil(float64(i+1) / float64(secondshardsize)))
		firstShard_leader[i] = (i + 1) % firstshardsize
		secondShard_leader[i] = (i + 1) % secondshardsize
	}
	nodelists := nodelist_shard{step1db.Nodes, firstShard, secondShard, firstShard_leader, secondShard_leader}
	sendByte := bytes.Buffer{}
	enc := gob.NewEncoder(&sendByte)
	err3 := enc.Encode(nodelists)
	if err3 != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("nodes_shard.gob", sendByte.Bytes(), 0666); err != nil {
		log.Fatal(err)
	}
	fmt.Println(nodelists)

}
