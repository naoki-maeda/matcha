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
	cfgFile string
	host string
	port string
	user string
	password string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "matcha",
	Short: "bitcoin developer tool",
	Long: `bitcoin developer tool`,
	RunE: run,
}

func run(cmd *cobra.Command, args []string) error {
	host = viper.GetString("host")
	port = viper.GetString("port")
	user = viper.GetString("user")
	password = viper.GetString("password")
	rpc, err := blockchain.NewRPC(host, port, user, password, true)
	if err != nil {
		return err
	}
	fmt.Println(rpc)
	fmt.Println(rpc.Client.GetBlockChainInfo())
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
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

	// Priority is cli default value < config file < cli args
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
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
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
