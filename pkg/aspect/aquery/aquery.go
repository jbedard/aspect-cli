/*
Copyright © 2021 Aspect Build Systems Inc

Not licensed for re-use.
*/

package aquery

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"aspect.build/cli/pkg/aspect/query/shared"
	"aspect.build/cli/pkg/bazel"
	"aspect.build/cli/pkg/ioutils"
)

type AQuery struct {
	ioutils.Streams

	Bzl           bazel.Bazel
	IsInteractive bool

	Presets []*shared.PresetQuery
	Prefs   viper.Viper

	Prompt       func(label string) shared.PromptRunner
	Confirmation func(question string) shared.ConfirmationRunner
	Select       func(presetNames []string) shared.SelectRunner
}

func New(streams ioutils.Streams, bzl bazel.Bazel, isInteractive bool) *AQuery {
	v := *viper.GetViper()

	presets := shared.GetPrecannedQueries("aquery", v)

	return &AQuery{
		Streams:       streams,
		Bzl:           bzl,
		IsInteractive: isInteractive,
		Presets:       presets,
		Prompt:        shared.Prompt,
		Select:        shared.Select,
		Confirmation:  shared.Confirmation,
		Prefs:         v,
	}
}

func (q *AQuery) Run(cmd *cobra.Command, args []string) error {
	verb := "aquery"
	presets, presetNames, err := shared.ProcessQueries(q.Presets)
	if err != nil {
		return shared.GetPrettyError(cmd, err)
	}

	verb, query, runReplacements, err := shared.SelectQuery(verb, presets, q.Presets, presetNames, q.Streams, args, q.Select)

	if err != nil {
		return shared.GetPrettyError(cmd, err)
	}

	if runReplacements {
		query, err = shared.ReplacePlaceholders(query, args, q.Prompt)

		if err != nil {
			return shared.GetPrettyError(cmd, err)
		}
	}

	return shared.RunQuery(q.Bzl, verb, query)
}
