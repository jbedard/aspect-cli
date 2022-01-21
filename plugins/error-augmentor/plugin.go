/*
Copyright © 2021 Aspect Build Systems Inc

Not licensed for re-use.
*/

package main

import (
	"fmt"
	"regexp"
	"strings"

	goplugin "github.com/hashicorp/go-plugin"
	"gopkg.in/yaml.v2"

	buildeventstream "aspect.build/cli/bazel/buildeventstream/proto"
	"aspect.build/cli/pkg/ioutils"
	"aspect.build/cli/pkg/plugin/sdk/v1alpha2/config"
)

func main() {
	goplugin.Serve(config.NewConfigFor(NewDefaultPlugin()))
}

type ErrorAugmentorPlugin struct {
	properties          string
	hintMap             map[*regexp.Regexp]string
	yamlUnmarshalStrict func(in []byte, out interface{}) (err error)
	helpfulHints        *helpfulHintSet
}

func NewDefaultPlugin() *ErrorAugmentorPlugin {
	return NewPlugin()
}

func NewPlugin() *ErrorAugmentorPlugin {
	return &ErrorAugmentorPlugin{
		properties:          "",
		hintMap:             map[*regexp.Regexp]string{},
		yamlUnmarshalStrict: yaml.UnmarshalStrict,
		helpfulHints:        &helpfulHintSet{nodes: make(map[helpfulHintNode]struct{})},
	}
}

// will be used to unmarshal plugin properties specific to this plugin
type errorMappings struct {
	ErrorMappings map[string]string `yaml:"error_mappings"`
}

func (plugin *ErrorAugmentorPlugin) SetupHook(
	properties string,
) error {
	plugin.properties = properties

	var processedProperties errorMappings
	if err := plugin.yamlUnmarshalStrict([]byte(properties), &processedProperties); err != nil {
		return fmt.Errorf("failed to setup: failed to parse properties: %w", err)
	}

	// change map keys into regex objects now so they are ready to use and we only need to compile the regex once
	for r, m := range processedProperties.ErrorMappings {
		// r for regex, m for message
		plugin.hintMap[regexp.MustCompile(r)] = m
	}

	return nil
}

func (plugin *ErrorAugmentorPlugin) BEPEventCallback(event *buildeventstream.BuildEvent) error {
	aborted := event.GetAborted()
	if aborted != nil {
		plugin.processErrorMessage(aborted.Description)

		// We exit early here because there will not be a progress message when the event was of type "aborted".
		return nil
	}

	progress := event.GetProgress()

	if progress != nil {
		stderr := progress.GetStderr()
		plugin.processErrorMessage(stderr)
	}

	return nil
}

func (plugin *ErrorAugmentorPlugin) processErrorMessage(errorMessage string) {
	for regex, helpfulHint := range plugin.hintMap {
		matches := regex.FindStringSubmatch(errorMessage)

		if len(matches) > 0 {
			str := fmt.Sprint(helpfulHint)

			for i, match := range matches {
				if i == 0 {
					continue
				}
				str = strings.ReplaceAll(helpfulHint, fmt.Sprint("$", i), match)
			}

			plugin.helpfulHints.insert(str)
		}
	}
}

func (plugin *ErrorAugmentorPlugin) PostBuildHook(
	isInteractiveMode bool,
	promptRunner ioutils.PromptRunner,
) error {
	if plugin.helpfulHints.size == 0 {
		return nil
	}

	plugin.printBreak()

	plugin.printMiddle("Aspect CLI Error Augmentor")
	plugin.printMiddle("")

	for node := plugin.helpfulHints.head; node != nil; node = node.next {
		plugin.printMiddle("- " + node.helpfulHint)
	}

	plugin.printBreak()
	return nil
}

func (plugin *ErrorAugmentorPlugin) printBreak() {
	var b strings.Builder

	fmt.Fprintf(&b, " ")

	for i := 0; i < 150; i++ {
		fmt.Fprintf(&b, "-")
	}

	fmt.Fprintf(&b, " ")

	fmt.Println(b.String())
}

func (plugin *ErrorAugmentorPlugin) printMiddle(str string) {
	var b strings.Builder

	fmt.Fprintf(&b, "| ")
	fmt.Fprintf(&b, str)

	for b.Len() < 151 {
		fmt.Fprintf(&b, " ")
	}

	fmt.Fprintf(&b, "|")
	fmt.Println(b.String())
}

func (plugin *ErrorAugmentorPlugin) PostTestHook(
	isInteractiveMode bool,
	promptRunner ioutils.PromptRunner,
) error {
	return plugin.PostBuildHook(isInteractiveMode, promptRunner)
}

func (plugin *ErrorAugmentorPlugin) PostRunHook(
	isInteractiveMode bool,
	promptRunner ioutils.PromptRunner,
) error {
	return plugin.PostBuildHook(isInteractiveMode, promptRunner)
}

type helpfulHintSet struct {
	head  *helpfulHintNode
	tail  *helpfulHintNode
	nodes map[helpfulHintNode]struct{}
	size  int
}

func (s *helpfulHintSet) insert(helpfulHint string) {
	node := helpfulHintNode{
		helpfulHint: helpfulHint,
	}
	if _, exists := s.nodes[node]; !exists {
		s.nodes[node] = struct{}{}
		if s.head == nil {
			s.head = &node
		} else {
			s.tail.next = &node
		}
		s.tail = &node
		s.size++
	}
}

type helpfulHintNode struct {
	next        *helpfulHintNode
	helpfulHint string
}
