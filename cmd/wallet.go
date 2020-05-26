package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
    bitSize int
)

// walletCmd represents the wallet command
var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Wallet Management command",
	Long: `Wallet Management command`,
	RunE: walletRun,
}

func walletRun(cmd *cobra.Command, args []string) error {
	fmt.Println("run wallet command")
	return nil
}

func init() {
	rootCmd.AddCommand(walletCmd)
	walletCmd.AddCommand(mnemonicCmd)
}
