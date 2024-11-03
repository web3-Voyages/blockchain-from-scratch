package cli

import (
	"blockchain-from-scratch/core"
	"blockchain-from-scratch/core/wallet"
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

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getbalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)

	getBalanceAddress := getbalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

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
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if getbalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getbalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}
}

func (cli *CLI) createBlockchain(address string) {
	bc := core.CreateBlockchain(address)
	defer bc.Db.Close()
	UTXOSet := core.UTXOSet{Blockchain: bc}
	UTXOSet.Reindex()
	fmt.Println("Done!")
}

func (cli *CLI) printChain() {
	bc := core.NewBlockChain()
	defer bc.Db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := core.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			utils.PrintJsonLog(tx, "printChain")
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) getBalance(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	chain := core.NewBlockChain()
	UTXOSet := core.UTXOSet{Blockchain: chain}
	defer chain.Db.Close()

	// The balance of a user's address is simply the sum of all UTXOs they own.
	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	fmt.Println(pubKeyHash)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	utxos := UTXOSet.FindUTXO(pubKeyHash)
	for _, out := range utxos {
		balance += out.Value
	}
	logrus.Infof("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	if !wallet.ValidateAddress(from) || !wallet.ValidateAddress(to) {
		log.Panic("ERROR: Address is not valid")
	}

	chain := core.NewBlockChain()
	UTXOSet := core.UTXOSet{Blockchain: chain}
	defer chain.Db.Close()

	tx := core.NewUTXOTransaction(from, to, amount, &UTXOSet)
	coinBaseRewardTx := core.NewCoinbaseTx(from, "")
	txs := []*core.Transaction{coinBaseRewardTx, tx}
	//utils.PrintJsonLog(txs, "send")
	newBlock := chain.MineBlock(txs)
	UTXOSet.Update(newBlock)
	fmt.Println("Success!")
}

func (cli *CLI) createWallet() {
	wallets, _ := wallet.NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a core and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the core")
	fmt.Println("  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
	fmt.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}
