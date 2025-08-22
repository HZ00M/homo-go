package main

import (
	"github.com/spf13/cobra"
	"log"

	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/change"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/project"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/proto"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/run"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/upgrade"
)

var rootCmd = &cobra.Command{
	Use:     "kratos",
	Short:   "Kratos: An elegant toolkit for Go microservices.",
	Long:    `Kratos: An elegant toolkit for Go microservices.`,
	Version: release,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Printf("rootCmd Executing command: %s %v", cmd.CommandPath(), args)
	},
}

func init() {
	rootCmd.AddCommand(project.CmdNew)
	rootCmd.AddCommand(proto.CmdProto)
	rootCmd.AddCommand(upgrade.CmdUpgrade)
	rootCmd.AddCommand(change.CmdChange)
	rootCmd.AddCommand(run.CmdRun)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
