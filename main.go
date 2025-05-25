package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type FileOrFolder struct {
	path string
	hash uint64
}

// type Config map[string]string
type Config struct {
	repo_link  string
	paths      []FileOrFolder
	generation uint32
}

/*
	rr, err := command.StderrPipe()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	i := 0
	lines := make([]string, 0, 10)
	err = command.Start()
	if err != nil {
		return nil, err
	}
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if callback != nil {
			callback(line, i)
		}
		i += 1
	}
	command.Wait()
	return lines, nil
*/

var CONFIG_PATH string

func run_cmd(app string, args string) (string, error) {
	command := exec.Command(app, strings.Split(args, " ")...)
	output, err := command.CombinedOutput()
	return string(output), err
}

func get_config() (Config, bool) {
	content, err := os.ReadFile(CONFIG_PATH + "config.toml")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// config file does not exist yet, show first time setup
			return Config{}, true
		}
	}
	var result Config
	_, err = toml.Decode(string(content), &result)
	return result, false
}

func save_config(config Config) {
	file, err := os.Create(CONFIG_PATH + "config.toml")
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(file)
	err = encoder.Encode(config)
	if err != nil {
		panic(err)
	}
	fmt.Println(buf.String())
	file.Close()
}

func show_loading(_ string, i int) {
	indicators := [...]string{"|", "/", "-", "\\"}
	fmt.Print("\rLoading " + indicators[i%4])
}

func onboarding(config Config) {
	for true {
		fmt.Println("To start syncing your files, you first need to create a git repository. When you're done, paste the link here:")
		var link string
		fmt.Scanln(&link)

		// test the repo link
		fmt.Println("Loading...")
		output, err := run_cmd("git", "clone --progress "+link+" "+CONFIG_PATH)
		_, ok := err.(*exec.ExitError)
		// if it is not an exit error
		if !ok && err != nil {
			fmt.Print("FATAL: Failed to clone the repository: ")
			fmt.Println(err)
			return
		}
		if strings.Contains(output, "fatal") {
			fmt.Println("Failed to clone the repository. Make sure that you are connected to the internet and that the following link was valid:\n" + link + "\n")
			fmt.Println("(", output, ")")
			fmt.Print("Try again? (y/n) ")
			var choice string
			fmt.Scanln(&choice)
			if choice != "y" && choice != "Y" {
				break
			} else {
				// if it is not an exit error
				continue
			}
		}

		// save the repo link
		config.repo_link = link
		break
	}
	fmt.Println("Done! You can now start adding files to track with `convene add <file/folder>` ")
}

func print_usage() {
	fmt.Println("Convene - tame and synchronize your scattered files with ease")
	fmt.Println("\nUsage:    convene [command] <args>")
	fmt.Println("\nThese are the available commands:")
	fmt.Println("    add [folder | file]   adds a folder/file to be synced")
	fmt.Println("    rm [folder | file]    stops tracking a folder/file")

	fmt.Println("\n    sync    synchronizes your files with the upstream repository")
	fmt.Println("\nTo get detailed Information, run `convene help [command]`.")
}

func CopyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}

func CopyFileOrDir(src string, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if sfi.Mode().IsRegular() { // if it is a file
		dfi, err := os.Stat(src)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		} else if os.SameFile(sfi, dfi) {
			return nil // if they are the same file, don't bother copying
		}
		return CopyFile(src, dst)
	} else if sfi.Mode().IsDir() {
		// copy files in the dir recursively
		fileCallback := func(path string, d DirEntry, err error) error {

		}
		filepath.WalkDir(src)
	}
}

func cmd_add(args []string) {

}
func cmd_rm(args []string) {
	fmt.Println("Remove is not implemented yet!")
}
func cmd_sync(args []string) {
	fmt.Println("Sync is not implemented yet!")
}

func cmd_help(args []string) {
	if len(args) == 0 {
		print_usage()
		return
	}
	switch args[0] {
	case "add":
		{
			fmt.Println("`add` tells Convene to track the provided file or folder. When calling `convene sync`, all tracked folders and files will be pushed to the remote repository. If the remote repository has newer changes or if there are any conflicts, you will be prompted.")
		}
	case "rm":
		{
			fmt.Println("`rm` tells Convene to stop tracking the provided file or folder. On the next call to `convene sync``, your files will be deleted from the remote repository.")
		}
	case "sync":
		{
			fmt.Println("`sync` tries to synchronize your state with the upstream repository. If the content of the remote repository is newer than your current state, your local files will be updated accordingly. If you made local changes, the remote repository will get updated too. Note that this command reuqires an internet connection.")
		}
	}
}

func main() {
	CONFIG_PATH = get_configpath()
	config, show_intro := get_config()
	// checking git version
	git_version, err := run_cmd("git", "--version")
	if err != nil {
		fmt.Println("Convene uses *git* to synchronize your files. Please make sure it is installed and on your PATH.")
		return
	}
	if show_intro || config.repo_link == "" {
		fmt.Println("# Welcome to Convene!")
		onboarding(config)
	}
	fmt.Println("- Using", git_version)

	switch os.Args[1] {
	case "add":
		{
			cmd_add(os.Args[2:])
		}
	case "rm":
		{
			cmd_rm(os.Args[2:])
		}
	case "sync":
		{
			cmd_sync(os.Args[2:])
		}
	case "help":
		{
			cmd_help(os.Args[2:])
		}
	default:
		{
			fmt.Println("Error: Unknown command '" + os.Args[1] + "'")
			print_usage()
		}
	}

	save_config(config)
}
