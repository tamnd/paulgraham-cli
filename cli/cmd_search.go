package cli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/tamnd/paulgraham-cli/paulgraham"
)

func (a *App) searchCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search essay titles (case-insensitive)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.ToLower(args[0])
			a.progressf("searching for %q...", args[0])
			essays, err := a.client.ListEssays(cmd.Context())
			if err != nil {
				return codeError(exitError, err)
			}

			var results []paulgraham.Essay
			for _, e := range essays {
				if strings.Contains(strings.ToLower(e.Title), query) {
					results = append(results, e)
				}
			}
			if limit > 0 && limit < len(results) {
				results = results[:limit]
			}
			return a.renderOrEmpty(results, len(results))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 10, "maximum number of results (0 = all)")
	return cmd
}
