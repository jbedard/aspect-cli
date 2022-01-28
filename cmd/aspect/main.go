/*
Copyright © 2021 Aspect Build Systems Inc

Not licensed for re-use.
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"aspect.build/cli/cmd/aspect/root"
	"aspect.build/cli/pkg/aspecterrors"
	"aspect.build/cli/pkg/ioutils"
	"aspect.build/cli/pkg/logger"
	"aspect.build/cli/pkg/plugin/system"
)

func main() {
	fmt.Println("In main")
	logger.Test1()
	logger.Log("Hello world from main.go")
	// Detect whether we are being run as a tools/bazel wrapper (look for BAZEL_REAL in the environment)
	// If so,
	//     Is this a bazel-native command? just call through to bazel without touching the arguments for now
	//     Is this an aspect-custom command? (like `outputs`) then write an implementation
	// otherwise,
	//     we are installing ourselves. Check with the user they intended to do that.
	//     then create
	//         - a WORKSPACE file, ask the user for the repository name if interactive
	//     ask the user if they want to install for all users of the workspace, if so
	//         - tools/bazel file and put our bootstrap code in there
	//

	// Convenience for local development: under `bazel run //:aspect` respect the
	// users working directory, don't run in the execroot
	if wd, exists := os.LookupEnv("BUILD_WORKING_DIRECTORY"); exists {
		_ = os.Chdir(wd)
	}

	pluginSystem := system.NewPluginSystem()
	if err := pluginSystem.Configure(ioutils.DefaultStreams); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	defer pluginSystem.TearDown()

	cmd := root.NewDefaultRootCmd(pluginSystem)
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		var exitErr *aspecterrors.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			os.Exit(exitErr.ExitCode)
		}

		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
