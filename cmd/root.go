package cmd

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/naoki-maeda/matcha/blockchain"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile        string
	host           string
	port           string
	user           string
	password       string
	network        string
	walletPassword string
	addressType    string
	mnemonic       string
	zmqAddress     string
	bitSize        int
	second         int
	addressCount   uint32
	verbose        bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "matcha",
	Short: "bitcoin developer tool",
	Long:  `bitcoin developer tool`,
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	// Priority is cli default value < config file < cli args
	host = viper.GetString("host")
	port = viper.GetString("port")
	user = viper.GetString("user")
	password = viper.GetString("password")
	network = viper.GetString("network")
	walletPassword = viper.GetString("wallet-password")
	addressType = viper.GetString("address-type")
	mnemonic = viper.GetString("mnemonic")
	zmqAddress = viper.GetString("zmq-address")
	bitSize = viper.GetInt("bit-size")
	second = viper.GetInt("second")
	addressCount = viper.GetUint32("address-count")
	verbose = viper.GetBool("verbose")

	rpc, err := blockchain.NewRPC(host, port, user, password, true)
	if err != nil {
		return err
	}

	hdwallet, err := blockchain.NewHDWallet(bitSize, mnemonic, network, walletPassword)
	if err != nil {
		return err
	}
	// write config file
	viper.Set("mnemonic", hdwallet.Mnemonic)
	if err := writeConfig(cfgFile); err != nil {
		return err
	}

	coinType := blockchain.GetCoinType(network)
	account, err := hdwallet.NewAccount(blockchain.Purpose, blockchain.CoinTypeBitcoinTestnet, coinType)
	if err != nil {
		return err
	}
	fmt.Println(`Available Accounts
==================`)
	addresses := make([]btcutil.Address, addressCount, addressCount)
	privKeys := make([]string, addressCount, addressCount)
	for i := uint32(0); i < addressCount; i++ {
		childWallet, err := account.DeriveAddress(blockchain.ChangeTypeExternal, i, addressType)
		if err != nil {
			return err
		}
		err = rpc.Client.ImportPrivKey(childWallet.WIF)
		if err != nil {
			return err
		}
		_, err = rpc.Client.GenerateToAddress(10, childWallet.Address, &blockchain.MaxTries)
		if err != nil {
			return err
		}
		fmt.Printf("(%d) %s\n", i, childWallet.Address)
		addresses[i] = childWallet.Address
		privKeys[i] = childWallet.WIF.String()
	}
	fmt.Println(`Private Keys
==================`)
	for n, privKey := range privKeys {
		fmt.Printf("(%d) %s\n", n, privKey)
	}

	fmt.Printf(
		`HD Wallet
==================
Mnemonic:      %s
Base HD Path:  m/44'/60'/0'/0/{account_index}
`, hdwallet.Mnemonic)

	if zmqAddress == "" {
		return nil
	}
	zmq, err := blockchain.NewZmqClient(zmqAddress, true)
	if err != nil {
		return err
	}
	msg := make(chan blockchain.Message)
	go zmq.Receive(msg)
	for {
		t := time.NewTicker(time.Duration(second) * time.Second)
		defer t.Stop()
		select {
		case <-t.C:
			// generate to random address
			rpc.Client.GenerateToAddress(1, addresses[rand.Intn(int(addressCount))], &blockchain.MaxTries)
		case message := <-msg:
			// maybe msg frames are fixed 3 length
			if message.Msg.Frames == nil || len(message.Msg.Frames) != 3 {
				break
			}
			topic := fmt.Sprintf("%s", message.Msg.Frames[0])
			switch topic {
			case blockchain.RawBlock:
				block, err := btcutil.NewBlockFromBytes(message.Msg.Frames[1])
				if err != nil {
					fmt.Println(err)
					return err
				}
				fmt.Println("==================")
				fmt.Printf("block hash: %s\n", block.Hash())
				if verbose == true {
					blockVerbose, err := rpc.Client.GetBlockVerbose(block.Hash())
					if err != nil {
						return err
					}
					j, err := json.Marshal(blockVerbose)
					if err != nil {
						return err
					}
					fmt.Println(string(j))
				}
			case blockchain.RawTx:
				tx, err := btcutil.NewTxFromBytes(message.Msg.Frames[1])
				if err != nil {
					return err
				}
				fmt.Println("==================")
				fmt.Printf("tx hash: %s\n", tx.Hash())
				if verbose == true {
					txVerbose, err := rpc.Client.GetRawTransactionVerbose(tx.Hash())
					if err != nil {
						return err
					}
					j, err := json.Marshal(txVerbose)
					if err != nil {
						return err
					}
					fmt.Println(string(j))
				}
			}
		}
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.SetOutput(os.Stdout)
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		// If exist error, output stderr
		rootCmd.SetOutput(os.Stderr)
		rootCmd.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().StringVar(&host, "host", "localhost", "bitcoind host name")
	rootCmd.PersistentFlags().StringVar(&port, "port", "18443", "bitcoind port")
	rootCmd.PersistentFlags().StringVar(&user, "user", "admin", "bitcoind user name")
	rootCmd.PersistentFlags().StringVar(&password, "password", "password", "bitcoind password")
	rootCmd.PersistentFlags().StringVar(&network, "network", "regtest", "bitcoind network")
	rootCmd.PersistentFlags().StringVar(&walletPassword, "wallet-password", "", "bitcoind HDWallet password")
	rootCmd.PersistentFlags().StringVar(&addressType, "address-type", "bech32", "bitcoin address type bech32 or p2kh or p2sh")
	rootCmd.PersistentFlags().StringVar(&mnemonic, "mnemonic", "", "HDWallet mnemonic")
	rootCmd.PersistentFlags().StringVar(&zmqAddress, "zmq-address", "tcp://localhost:28332", "zero mq address")
	rootCmd.PersistentFlags().IntVar(&bitSize, "bit-size", 128, "bit-size must be [128, 256] and a multiple of 32")
	rootCmd.PersistentFlags().IntVar(&second, "second", 30, "generate block automatically by second")
	rootCmd.PersistentFlags().Uint32Var(&addressCount, "address-count", 10, "generate and import bitcoin address count")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", true, "Print bitcoind block and tx verbose")

	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("network", rootCmd.PersistentFlags().Lookup("network"))
	viper.BindPFlag("wallet-password", rootCmd.PersistentFlags().Lookup("wallet-password"))
	viper.BindPFlag("address-type", rootCmd.PersistentFlags().Lookup("address-type"))
	viper.BindPFlag("mnemonic", rootCmd.PersistentFlags().Lookup("mnemonic"))
	viper.BindPFlag("zmq-address", rootCmd.PersistentFlags().Lookup("zmq-address"))
	viper.BindPFlag("bit-size", rootCmd.PersistentFlags().Lookup("bit-size"))
	viper.BindPFlag("second", rootCmd.PersistentFlags().Lookup("second"))
	viper.BindPFlag("address-count", rootCmd.PersistentFlags().Lookup("address-count"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file if set.
// ENV variables disabled
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Search config in home directory with name ".matcha" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".matcha")
		viper.SetConfigType("yaml")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func writeConfig(configFile string) error {
	if configFile != "" {
		err := viper.WriteConfigAs(configFile)
		return err
	}
	err := viper.WriteConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			home, err := homedir.Dir()
			if err != nil {
				return err
			}
			return viper.WriteConfigAs(home + "/.matcha.yaml")
		}
	}
	return err
}
