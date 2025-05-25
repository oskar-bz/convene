//go:build windows
// +build windows

package main

import "os"

func get_configpath() string {
	return os.Getenv("APPDATA") + "\\convene\\"
}
