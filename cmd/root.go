package cmd

import (
	"fmt"
	"os"

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
	size           int
	walletPassword string
	addressType    string
	mnemonic       string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "matcha",
	Short: "bitcoin developer tool",
	Long:  `bitcoin developer tool`,
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	host = viper.GetString("host")
	port = viper.GetString("port")
	user = viper.GetString("user")
	password = viper.GetString("password")
	network = viper.GetString("network")
	size = viper.GetInt("size")
	walletPassword = viper.GetString("wallet-password")
	addressType = viper.GetString("address-type")
	mnemonic = viper.GetString("mnemonic")

	rpc, err := blockchain.NewRPC(host, port, user, password, true)
	if err != nil {
		return err
	}
	blockInfo, err := rpc.Client.GetBlockChainInfo()
	if err != nil {
		return err
	}
	fmt.Println(blockInfo)

	hdwallet, err := blockchain.NewHDWallet(size, mnemonic, network, walletPassword)
	if err != nil {
		return err
	}
	// write config file
	viper.Set("mnemonic", hdwallet.Mnemonic)

	coinType := blockchain.GetCoinType(network)
	account, err := hdwallet.NewAccount(blockchain.Purpose, blockchain.CoinTypeBitcoinTestnet, coinType)
	if err != nil {
		return err
	}
	fmt.Println(`Available Accounts
==================`)
	var privKeys [10]string
	for i := uint32(0); i < 10; i++ {
		childWallet, err := account.DeriveAddress(blockchain.ChangeTypeExternal, i, addressType)
		if err != nil {
			return err
		}
		fmt.Printf("(%d) %s\n", i, childWallet.Address)
		privKeys[i] = childWallet.PrivKey
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
Base HD Path:  m/44'/60'/0'/0/{account_index}`, hdwallet.Mnemonic)
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// write config file
	if err := viper.WriteConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.matcha.yaml)")
	rootCmd.PersistentFlags().StringVar(&host, "host", "localhost", "bitcoind host name (default is localhost")
	rootCmd.PersistentFlags().StringVar(&port, "port", "18443", "bitcoind port (default is regtest port)")
	rootCmd.PersistentFlags().StringVar(&user, "user", "admin", "bitcoind user name (default is admin)")
	rootCmd.PersistentFlags().StringVar(&password, "password", "password", "bitcoind password (default is password)")
	rootCmd.PersistentFlags().StringVar(&network, "network", "regtest", "bitcoind network (default is regtest)")
	rootCmd.PersistentFlags().IntVar(&size, "size", 128, "bitSize must be [128, 256] and a multiple of 32 (default is 128)")
	rootCmd.PersistentFlags().StringVar(&walletPassword, "wallet-password", "", "bitcoind HDWallet password (default is nothing)")
	rootCmd.PersistentFlags().StringVar(&addressType, "address-type", "bech32", "bitcoin address type bech32 or p2kh or p2sh (default is bech32)")
	rootCmd.PersistentFlags().StringVar(&mnemonic, "mnemonic", "", "HDWallet mnemonic")

	// Priority is cli default value < config file < cli args
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("network", rootCmd.PersistentFlags().Lookup("network"))
	viper.BindPFlag("size", rootCmd.PersistentFlags().Lookup("size"))
	viper.BindPFlag("wallet-password", rootCmd.PersistentFlags().Lookup("wallet-password"))
	viper.BindPFlag("address-type", rootCmd.PersistentFlags().Lookup("address-type"))
	viper.BindPFlag("mnemonic", rootCmd.PersistentFlags().Lookup("mnemonic"))
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
