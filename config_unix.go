//go:build unix
// +build unix

package main

import "os"

func get_configpath() string {
	return os.Getenv("HOME") + "/.config/convene/"
}
