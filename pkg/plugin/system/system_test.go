/*
Copyright © 2022 Aspect Build Systems Inc

Not licensed for re-use.
*/

package system

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	rootFlags "aspect.build/cli/pkg/aspect/root/flags"
	"aspect.build/cli/pkg/aspecterrors"
	"aspect.build/cli/pkg/ioutils"
	plugin_mock "aspect.build/cli/pkg/plugin/sdk/v1alpha2/plugin/mock"
)

func TestPluginSystem(t *testing.T) {
	t.Run("executes hooks in reverse order of interceptors", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Minimal streams
		var stdout strings.Builder
		streams := ioutils.Streams{Stdout: &stdout, Stderr: &stdout}

		// empty/mock ctx, prompt-runner, command
		ctx := context.Background()
		promptRunner := ioutils.NewPromptRunner()
		cmd := &cobra.Command{
			Use: "HookOrderTest",
		}
		// Required flags for interceptor hooks
		cmd.PersistentFlags().Bool(rootFlags.InteractiveFlagName, false, "")

		// Mock plugin
		plugin := plugin_mock.NewMockPlugin(ctrl)
		plugins := PluginList{}
		plugins.insert(plugin)

		// Expect the callbacks in reverse-order of execution
		gomock.InOrder(
			plugin.EXPECT().PostRunHook(gomock.Any(), gomock.Any()),
			plugin.EXPECT().PostTestHook(gomock.Any(), gomock.Any()),
			plugin.EXPECT().PostBuildHook(gomock.Any(), gomock.Any()),
		)

		// Hook interceptors
		buildInterceptor := commandHooksInterceptor(&plugins, promptRunner, "PostBuildHook", streams)
		testInterceptor := commandHooksInterceptor(&plugins, promptRunner, "PostTestHook", streams)
		runInterceptor := commandHooksInterceptor(&plugins, promptRunner, "PostRunHook", streams)

		err := buildInterceptor(ctx, cmd, []string{}, func(ctx context.Context, cmd *cobra.Command, args []string) error {
			return testInterceptor(ctx, cmd, args, func(ctx context.Context, cmd *cobra.Command, args []string) error {
				return runInterceptor(ctx, cmd, args, func(ctx context.Context, cmd *cobra.Command, args []string) error {
					return nil
				})
			})
		})

		g.Expect(err).To(BeNil())
	})

	t.Run("executes plugin hooks in order plugins are added", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Minimal streams
		var stdout strings.Builder
		streams := ioutils.Streams{Stdout: &stdout, Stderr: &stdout}

		// empty/mock ctx, prompt-runner, command
		ctx := context.Background()
		promptRunner := ioutils.NewPromptRunner()
		cmd := &cobra.Command{
			Use: "HookOrderTest",
		}
		// Required flags for interceptor hooks
		cmd.PersistentFlags().Bool(rootFlags.InteractiveFlagName, false, "")

		plugin1 := plugin_mock.NewMockPlugin(ctrl)
		plugin2 := plugin_mock.NewMockPlugin(ctrl)

		// Expect the callbacks in reverse-order of execution, plugins in order added
		gomock.InOrder(
			plugin1.EXPECT().PostTestHook(gomock.Any(), gomock.Any()),
			plugin2.EXPECT().PostTestHook(gomock.Any(), gomock.Any()),
			plugin1.EXPECT().PostBuildHook(gomock.Any(), gomock.Any()),
			plugin2.EXPECT().PostBuildHook(gomock.Any(), gomock.Any()),
		)

		plugins := PluginList{}
		plugins.insert(plugin1)
		plugins.insert(plugin2)

		// Hook interceptors
		buildInterceptor := commandHooksInterceptor(&plugins, promptRunner, "PostBuildHook", streams)
		testInterceptor := commandHooksInterceptor(&plugins, promptRunner, "PostTestHook", streams)

		err := buildInterceptor(ctx, cmd, []string{}, func(ctx context.Context, cmd *cobra.Command, args []string) error {
			return testInterceptor(ctx, cmd, args, func(ctx context.Context, cmd *cobra.Command, args []string) error {
				return nil
			})
		})

		g.Expect(err).To(BeNil())
	})

	t.Run("returns pass nested interceptor errors to parent", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Minimal streams
		var stdout strings.Builder
		streams := ioutils.Streams{Stdout: &stdout, Stderr: &stdout}

		// empty/mock ctx, prompt-runner, command
		ctx := context.Background()
		promptRunner := ioutils.NewPromptRunner()
		cmd := &cobra.Command{
			Use: "HookOrderTest",
		}
		// Required flags for interceptor hooks
		cmd.PersistentFlags().Bool(rootFlags.InteractiveFlagName, false, "")

		plugin := plugin_mock.NewMockPlugin(ctrl)

		// Expect the callbacks in reverse-order of execution
		gomock.InOrder(
			plugin.EXPECT().PostRunHook(gomock.Any(), gomock.Any()),
			plugin.EXPECT().PostTestHook(gomock.Any(), gomock.Any()),
			plugin.EXPECT().PostBuildHook(gomock.Any(), gomock.Any()),
		)

		plugins := PluginList{}
		plugins.insert(plugin)

		// Hook interceptors
		buildInterceptor := commandHooksInterceptor(&plugins, promptRunner, "PostBuildHook", streams)
		testInterceptor := commandHooksInterceptor(&plugins, promptRunner, "PostTestHook", streams)
		runInterceptor := commandHooksInterceptor(&plugins, promptRunner, "PostRunHook", streams)

		err := buildInterceptor(ctx, cmd, []string{}, func(ctx context.Context, cmd *cobra.Command, args []string) error {
			return testInterceptor(ctx, cmd, args, func(ctx context.Context, cmd *cobra.Command, args []string) error {
				return runInterceptor(ctx, cmd, args, func(ctx context.Context, cmd *cobra.Command, args []string) error {
					return fmt.Errorf("test error")
				})
			})
		})

		g.Expect(err).To(MatchError("test error"))
	})

	t.Run("parent interceptor errors override child errors", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Minimal streams
		var stdout strings.Builder
		streams := ioutils.Streams{Stdout: &stdout, Stderr: &stdout}

		// empty/mock ctx, prompt-runner, command
		ctx := context.Background()
		promptRunner := ioutils.NewPromptRunner()
		cmd := &cobra.Command{
			Use: "HookOrderTest",
		}
		// Required flags for interceptor hooks
		cmd.PersistentFlags().Bool(rootFlags.InteractiveFlagName, false, "")

		plugin := plugin_mock.NewMockPlugin(ctrl)

		// Expect the callbacks in reverse-order of execution
		gomock.InOrder(
			plugin.EXPECT().PostRunHook(gomock.Any(), gomock.Any()),
			plugin.EXPECT().PostTestHook(gomock.Any(), gomock.Any()),
			plugin.EXPECT().PostBuildHook(gomock.Any(), gomock.Any()),
		)

		plugins := PluginList{}
		plugins.insert(plugin)

		// Hook interceptors
		buildInterceptor := commandHooksInterceptor(&plugins, promptRunner, "PostBuildHook", streams)
		testInterceptor := commandHooksInterceptor(&plugins, promptRunner, "PostTestHook", streams)
		runInterceptor := commandHooksInterceptor(&plugins, promptRunner, "PostRunHook", streams)

		err := buildInterceptor(ctx, cmd, []string{}, func(ctx context.Context, cmd *cobra.Command, args []string) error {
			return testInterceptor(ctx, cmd, args, func(ctx context.Context, cmd *cobra.Command, args []string) error {
				runInterceptor(ctx, cmd, args, func(ctx context.Context, cmd *cobra.Command, args []string) error {
					return fmt.Errorf("error 1")
				})
				return fmt.Errorf("error 2")
			})
		})

		g.Expect(err).To(MatchError("error 2"))
	})

	t.Run("should set ExitCode=1 on error of type aspecterrors.ExitError", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Minimal streams
		var stdout strings.Builder
		streams := ioutils.Streams{Stdout: &stdout, Stderr: &stdout}

		// empty/mock ctx, prompt-runner, command
		ctx := context.Background()
		promptRunner := ioutils.NewPromptRunner()
		cmd := &cobra.Command{
			Use: "HookOrderTest",
		}
		// Required flags for interceptor hooks
		cmd.PersistentFlags().Bool(rootFlags.InteractiveFlagName, false, "")

		plugin := plugin_mock.NewMockPlugin(ctrl)
		plugin.EXPECT().
			PostRunHook(gomock.Any(), gomock.Any()).
			DoAndReturn(func(
				isInteractiveMode bool,
				promptRunner ioutils.PromptRunner,
			) interface{} {
				return &aspecterrors.ExitError{
					Err:      fmt.Errorf("plugin error"),
					ExitCode: 123,
				}
			})
		plugins := PluginList{}
		plugins.insert(plugin)

		// Hook interceptors
		runInterceptor := commandHooksInterceptor(&plugins, promptRunner, "PostRunHook", streams)

		err := runInterceptor(ctx, cmd, []string{}, func(ctx context.Context, cmd *cobra.Command, args []string) error {
			return nil
		})

		g.Expect(err.(*aspecterrors.ExitError).Err).To(MatchError("aspect error"))
		g.Expect(err.(*aspecterrors.ExitError).ExitCode).To(Equal(1))
	})
}
