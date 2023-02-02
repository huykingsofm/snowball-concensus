package main

import (
	"github.com/huykingsofm/snowball-concensus/internal/consensus"
	"github.com/spf13/cobra"
)

var nTx, kSamples, alpha, beta uint
var folder string
var port, minPort, maxPort int

var rootCommand = &cobra.Command{
	Use: "consensus",
	Run: func(cmd *cobra.Command, args []string) {
		app := consensus.New(folder, port, minPort, maxPort, nTx, kSamples, alpha, beta)
		app.Run()
	},
}

func main() {
	rootCommand.Flags().IntVar(&port, "p", 0, "the listenning port")
	rootCommand.Flags().IntVar(&minPort, "minP", 60000, "the minimum peer port")
	rootCommand.Flags().IntVar(&maxPort, "maxP", 60200, "the maximum peer port")
	rootCommand.Flags().StringVar(&folder, "folder", "storage", "the folder contains final results")
	rootCommand.Flags().UintVar(&nTx, "n", 10, "number of transactions")
	rootCommand.Flags().UintVar(&kSamples, "k", 4, "number of samples")
	rootCommand.Flags().UintVar(&alpha, "a", 3, "minimum number of consensus")
	rootCommand.Flags().UintVar(&beta, "b", 5, "minimum number of consecutive success")
	if err := rootCommand.Execute(); err != nil {
		panic(err)
	}
}
