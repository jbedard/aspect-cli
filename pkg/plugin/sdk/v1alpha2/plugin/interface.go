/*
Copyright © 2021 Aspect Build Systems Inc

Not licensed for re-use.
*/

package plugin

import (
	buildeventstream "aspect.build/cli/bazel/buildeventstream/proto"
	"aspect.build/cli/pkg/ioutils"
)

// Plugin determines how an aspect Plugin should be implemented.
type Plugin interface {
	BEPEventCallback(event *buildeventstream.BuildEvent) error
	PostBuildHook(
		isInteractiveMode bool,
		promptRunner ioutils.PromptRunner,
	) error
	CustomCommands(
		isInteractiveMode bool,
		promptRunner ioutils.PromptRunner,
	) []*Command
	PostTestHook(
		isInteractiveMode bool,
		promptRunner ioutils.PromptRunner,
	) error
	PostRunHook(
		isInteractiveMode bool,
		promptRunner ioutils.PromptRunner,
	) error
}

type Command struct {
	Use string
	Run func(args []string) error
	// Run func(cmd *Command, args []string) error
}
