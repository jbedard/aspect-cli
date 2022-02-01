//go:build tools
// +build tools

// Libraries required by rules_go which we do not directly use which causes go tidy to remove it

package aspect

import _ "golang.org/x/tools/godoc"
