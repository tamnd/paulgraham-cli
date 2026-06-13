package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) listCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all Paul Graham essays",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a.progressf("fetching essays list...")
			essays, err := a.client.ListEssays(cmd.Context())
			if err != nil {
				return codeError(exitError, err)
			}
			if limit > 0 && limit < len(essays) {
				essays = essays[:limit]
			}
			return a.renderOrEmpty(essays, len(essays))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum number of essays to show (0 = all)")
	return cmd
}
