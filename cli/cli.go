package cli

import (
	"blockchain-from-scratch/core"
	"blockchain-from-scratch/core/wallet"
	"blockchain-from-scratch/node"
	"blockchain-from-scratch/utils"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

type CLI struct {
	Chain *core.Blockchain
}

func (cli *CLI) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env. var is not set!")
		os.Exit(1)
	}

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getbalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	getUTXODetailsCmd := flag.NewFlagSet("getUTXODetails", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getBalanceAddress := getbalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	case "getbalance":
		err := getbalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getUTXODetails":
		err := getUTXODetailsCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress, nodeID)
	}

	if getbalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getbalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}
	if getUTXODetailsCmd.Parsed() {
		cli.GetUTXODetails(nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, nodeID, *sendAmount, *sendMine)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}

	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			os.Exit(1)
		}
		cli.startNode(nodeID, *startNodeMiner)
	}
}

func (cli *CLI) createBlockchain(address, nodeID string) {
	bc := core.CreateBlockchain(address, nodeID)
	defer bc.Db.Close()
	UTXOSet := core.UTXOSet{Blockchain: bc}
	UTXOSet.Reindex()
	fmt.Println("Done!")
}

func (cli *CLI) printChain(nodeID string) {
	bc := core.NewBlockChain(nodeID)
	defer bc.Db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()
		//fmt.Printf("============ Block %x ============\n", block.Hash)
		//fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := core.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		//for _, tx := range block.Transactions {
		//	utils.PrintJsonLog(tx, "printChain")
		//}
		utils.PrintJsonLog(block, "Block")
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) getBalance(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	chain := core.NewBlockChain(nodeID)
	UTXOSet := core.UTXOSet{Blockchain: chain}
	defer chain.Db.Close()

	// The balance of a user's address is simply the sum of all UTXOs they own.
	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	utxos := UTXOSet.FindUTXO(pubKeyHash)
	for _, out := range utxos {
		balance += out.Value
	}
	logrus.Infof("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) send(from, to, nodeID string, amount int, mineNow bool) {
	if !wallet.ValidateAddress(from) || !wallet.ValidateAddress(to) {
		log.Panic("ERROR: Address is not valid")
	}

	chain := core.NewBlockChain(nodeID)
	UTXOSet := core.UTXOSet{Blockchain: chain}
	defer chain.Db.Close()

	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := core.NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	if mineNow {
		cbTx := core.NewCoinbaseTx(from, "")
		txs := []*core.Transaction{cbTx, tx}

		newBlock := chain.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		node.SendTxToNode(tx)
	}

	fmt.Println("Success!")
}

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := wallet.NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) GetUTXODetails(nodeID string) {
	chain := core.NewBlockChain(nodeID)
	UTXOSet := core.UTXOSet{Blockchain: chain}
	UTXOSet.GetUTXODetails()
	//UTXOSet.Reindex()
	//UTXOSet.GetUTXODetails()
}

func (cli *CLI) startNode(nodeID, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	node.StartServer(nodeID, minerAddress)
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a core and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the core")
	fmt.Println("  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("  getUTXODetails - Get UTXO Set details")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
	fmt.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}
