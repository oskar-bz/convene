//go:build windows
// +build windows

package main

import "os"
import "strings"

const SEPERATOR = "\\"
const NOT_SEPERATOR = "/"

func get_configpath() string {
	return os.Getenv("APPDATA") + "\\convene\\"
}

func GetPathReplacers() map[string]string {
	result := make(map[string]string)
	result[strings.ToLower(os.Getenv("APPDATA"))+"\\"] = "%appdata%\\"
	result[strings.ToLower(os.Getenv("USERPROFILE"))+"\\"] = "%userprofile%\\"
	result[strings.ToLower(os.Getenv("ALLUSERSPROFILE"))+"\\"] = "%ALLUSERSPROFILE%\\"
	result[strings.ToLower(os.Getenv("COMMONPROGRAMFILES"))+"\\"] = "%COMMONPROGRAMFILES%\\"
	result[strings.ToLower(os.Getenv("PROGRAMFILES"))+"\\"] = "%PROGRAMFILES%\\"
	result[strings.ToLower(os.Getenv("PROGRAMFILES(X86)"))+"\\"] = "%PROGRAMFILES(X86)%\\"
	result[strings.ToLower(os.Getenv("SYSTEMDRIVE"))+"\\"] = "%SYSTEMDRIVE%\\"
	result[strings.ToLower(os.Getenv("SYSTEMROOT"))+"\\"] = "%SYSTEMROOT%\\"
	result[strings.ToLower(os.Getenv("APPDATA"))+"\\"] = "%APPDATA%\\"
	result[strings.ToLower(os.Getenv("LOCALAPPDATA"))+"\\"] = "%LOCALAPPDATA%\\"
	result[strings.ToLower(os.Getenv("HOMEPATH"))+"\\"] = "%HOMEPATH%\\"
	result[strings.ToLower(os.Getenv("TEMP"))+"\\"] = "%TEMP%\\"
	result[strings.ToLower(os.Getenv("TMP"))+"\\"] = "%TMP%\\"
	result[strings.ToLower(os.Getenv("USERPROFILE"))+"\\"] = "%USERPROFILE%\\"
	return result
}
