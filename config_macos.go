//go:build darwin
// +build darwin

package main

import "os"

func get_configpath() string {
	return os.Getenv("HOME") + "/Library/Application Support"
}
