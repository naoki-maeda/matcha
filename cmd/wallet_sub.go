package cmd

import (
	"fmt"

	"github.com/naoki-maeda/matcha/blockchain"
	"github.com/spf13/cobra"
)

// mnemonicCmd generate mnemonic by bitSize
var mnemonicCmd = &cobra.Command{
	Use:   "mnemonic",
	Short: "generate mnemonic",
	Long:  `generate mnemonic by bitSize`,
	RunE:  generateMnemonic,
}

func generateMnemonic(cmd *cobra.Command, args []string) error {
	mnemonic, err := blockchain.GenerateMnemonic(bitSize)
	if err != nil {
		return err
	}
	fmt.Println(mnemonic)
	return nil
}

func init() {
	mnemonicCmd.PersistentFlags().IntVarP(&bitSize, "size", "s", 128, "bitSize must be [128, 256] and a multiple of 32")
}
