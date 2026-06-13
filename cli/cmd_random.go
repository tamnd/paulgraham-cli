package cli

import (
	"math/rand"

	"github.com/spf13/cobra"
)

func (a *App) randomCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "random",
		Short: "Show a random Paul Graham essay",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a.progressf("fetching essays list...")
			essays, err := a.client.ListEssays(cmd.Context())
			if err != nil {
				return codeError(exitError, err)
			}
			if len(essays) == 0 {
				return codeError(exitNoData, nil)
			}
			picked := essays[rand.Intn(len(essays))]
			a.progressf("fetching essay %q...", picked.Slug)
			content, err := a.client.GetEssay(cmd.Context(), picked.Slug)
			if err != nil {
				return codeError(exitError, err)
			}
			// Default to raw for essay content
			if a.output == string(FormatTable) {
				a.output = string(FormatRaw)
			}
			return a.render(content)
		},
	}
}
