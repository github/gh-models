package view

import "github.com/spf13/cobra"

func NewViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View details about a model",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}
