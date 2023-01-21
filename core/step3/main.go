package main

import (
	"fmt"
)

var Core = struct {
	*Keypair
	*Blockchain
	*Network
}{}

func Start(address string) {

	// Setup keys
	keypair, _ := OpenConfiguration(HOME_DIRECTORY_CONFIG)
	if keypair == nil {

		fmt.Println("Generating keypair...")
		keypair = GenerateNewKeypair()
		WriteConfiguration(HOME_DIRECTORY_CONFIG, keypair)
	}
	Core.Keypair = keypair

	// Setup Network
	Core.Network = SetupNetwork(address)
	go Core.Network.Run()
	node_info1, node_info2 := readNodesinfo("127.0.0.1:7770")
	Core.Network.connecttonode(node_info1, node_info2)

	// Setup blockchain
	Core.Blockchain = SetupBlockchain()
	go Core.Blockchain.Run()

	go func() {
		for {
			select {
			case msg := <-Core.Network.IncomingMessages:
				HandleIncomingMessage(msg)
			}
		}
	}()
}
