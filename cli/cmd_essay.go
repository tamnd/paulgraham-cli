package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) essayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "essay <slug>",
		Short: "Show a specific essay by slug",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			a.progressf("fetching essay %q...", slug)
			content, err := a.client.GetEssay(cmd.Context(), slug)
			if err != nil {
				return codeError(exitError, err)
			}
			// Default to raw for essay content so body prints in full.
			if a.output == string(FormatTable) {
				a.output = string(FormatRaw)
			}
			return a.render(content)
		},
	}
}
