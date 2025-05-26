package main

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type FileOrFolder struct {
	Hash [md5.Size]byte
}

type Config struct {
	RepoLink    string
	Paths       map[string]FileOrFolder
	SyncedPaths map[string]string
	Generation  uint32
}

var CONFIG_PATH string
var PATH_REPLACERS map[string]string
var PATH_REPLACERS_SORTED []string

func run_cmd(app string, args string) (string, error) {
	command := exec.Command(app, strings.Split(args, " ")...)
	output, err := command.CombinedOutput()
	return string(output), err
}

func run_cmd_in(app string, args string, wd string) (string, error) {
	command := exec.Command(app, strings.Split(args, " ")...)
	command.Dir = CONFIG_PATH
	output, err := command.CombinedOutput()
	return string(output), err
}

func get_config() (Config, bool) {
	content, err := os.ReadFile(CONFIG_PATH + "config.Toml")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// config file does not exist yet, show first time setup
			return Config{}, true
		}
	}
	var result Config
	_, err = toml.Decode(string(content), &result)
	_ = err

	if result.Paths == nil {
		result.Paths = make(map[string]FileOrFolder)
	}
	return result, false
}

func save_config(config *Config) {
	file, err := os.Create(CONFIG_PATH + "config.Toml")
	if err != nil {
		panic(err)
	}

	encoder := toml.NewEncoder(file)
	err = encoder.Encode(&config)
	if err != nil {
		panic(err)
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}
}

func show_loading(i int) {
	indicators := [...]string{"|", "/", "-", "\\"}
	i /= 4
	fmt.Print("\rLoading " + indicators[i%4])
}

func onboarding(config *Config) {
	for {
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
		config.RepoLink = link
		break
	}
	fmt.Println("Done! You can now start adding files to track with `convene add <file/folder>` ")
}

func print_usage() {
	fmt.Println("Convene - tame and synchronize your scattered files with ease")
	fmt.Println("\nUsage: convene [command] <args>")
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

func SplitFileFromPath(path string) (string, string) {
	split_at := 0
	for i := len(path) - 1; i > 0; i-- {
		c := path[i]
		if c == '/' || c == '\\' {
			split_at = i
			break
		}
	}
	return path[:split_at+1], path[split_at+1:]
}

func MovePath(src string, dst string, base string) string {
	rel_path := strings.Replace(src, base, "", 1)
	return dst + rel_path
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
		fileCallback := func(path string, d fs.DirEntry, err error) error {
			if !d.Type().IsRegular() {
				return nil
			}
			new_path := MovePath(path, dst, src)
			CopyFile(path, new_path)
			return nil
		}
		err := filepath.WalkDir(src, fileCallback)
		return err
	}
	return nil
}

func HashFile(path string) ([md5.Size]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return [md5.Size]byte{}, err
	}
	defer file.Close()

	// Get the file size
	stat, err := file.Stat()
	if err != nil {
		return [16]byte{}, err
	}
	// Read the file into a byte slice
	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(bs)
	if err != nil && err != io.EOF {
		return [16]byte{}, err
	}

	return md5.Sum(bs), nil
}

func HashPath(path string) ([md5.Size]byte, error) {
	s, err := os.Stat(path)
	if err != nil {
		return [md5.Size]byte{}, err
	}

	if s.Mode().IsRegular() {
		// if it is a file
		return HashFile(path)
	}
	// if it is a directory
	hashes := make([]byte, 0, 16*10)
	i := 0
	callback := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.Contains(path, "/.") || strings.Contains(path, "\\.") {
			return nil
		}

		if !d.Type().IsRegular() {
			return nil
		}

		h, err := HashFile(path)
		if err != nil {
			fmt.Printf("Failed to Add file %s (%s)\n", path, err.Error())
			return nil
		}
		//dots := [3]string{".", "..", "..."}
		fmt.Printf("\rLoading %d", i)
		//show_loading(i)
		i += 1
		for j := range md5.Size {
			hashes = append(hashes, h[j])
		}
		return nil
	}
	err = filepath.WalkDir(path, callback)

	return md5.Sum(hashes), err
}

func NormalizePath(path string) string {
	path = strings.ToLower(path)
	path = strings.Replace(path, NOT_SEPERATOR, SEPERATOR, -1)

	//fmt.Println("hey", PATH_REPLACERS_SORTED)
	for _, key := range PATH_REPLACERS_SORTED {
		value := PATH_REPLACERS[key]
		//fmt.Println("Trying", key, "=>", value, "in:", path)
		path = strings.Replace(path, key, value, 1)
	}
	if path[len(path)-1] == SEPERATOR[0] {
		path = path[:len(path)-1]
	}
	return path
}

func cmd_add(config *Config, args []string) {
	if len(args) == 0 {
		fmt.Println("Nothing to add.")
		return
	}

	for _, arg := range args {
		// check if path is already in tracked paths
		arg, err := filepath.Abs(arg)
		if err != nil {
			fmt.Printf("Error: Could not add '%s' (%s)", arg, err.Error())
			continue
		}
		found := false
		for key, _ := range config.Paths {
			if strings.Contains(arg, key) {
				found = true
			}
		}
		if found {
			fmt.Printf("'%s' is already tracked\n", arg)
			continue
		}

		// normalize path
		normalized := NormalizePath(arg)
		_, err = os.Stat(arg)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("Error: Could not add '%s', because it does not exist\n", arg)
			} else {
				fmt.Printf("Error: Could not add '%s' (%s)\n", arg, err.Error())
			}
			continue
		}
		h, err := HashPath(arg)
		if err != nil {
			fmt.Printf("Error: Could not add '%s' (%s)\n", arg, err.Error())
			continue
		}
		fmt.Printf("\rAdded '%s'\n", normalized)
		config.Paths[normalized] = FileOrFolder{Hash: h}
	}
}

func cmd_rm(config *Config, args []string) {
	if len(args) == 0 {
		fmt.Println("Nothing to remove.")
		return
	}
	for _, arg := range args {
		arg, err := filepath.Abs(arg)
		if err != nil {
			fmt.Printf("Error: Could not remove '%s' (%s)", arg, err.Error())
			continue
		}
		arg = NormalizePath(arg)
		_, ok := config.Paths[arg]
		if !ok {
			fmt.Println("Error: Could not remove", arg, "because it was not tracked")
			continue
		}
		delete(config.Paths, arg)
		fmt.Println("Removed", arg)
	}
}

func cmd_list(config *Config) {
	fmt.Println("Tracked Paths:")
	for key, _ := range config.Paths {
		fmt.Println("-", key)
	}
}

func cmd_sync(config *Config, args []string) {
	// run git pull
	out, err := run_cmd_in("git", "pull", CONFIG_PATH)
	if err != nil {
		fmt.Printf("Failed to sync: %s\n", err.Error())
		return
	}
	if strings.Contains(out, "up to date") {
		// if there were no upstream changes
		i := 0
		for key, _ := range config.Paths {
			end, front := SplitFileFromPath(key)
			config.SyncedPaths[end] = front
			CopyFileOrDir(key, CONFIG_PATH+end+SEPERATOR)
			fmt.Println("Loading", i)
			i += 1
		}
		output, err := run_cmd_in("git", "commit -m '"+time.Now().Local().Format(time.RFC1123), "'")
		if err != nil {
			fmt.Printf("Error: Failed to commit data (%s)", err.Error())
		}
		return
	}
	changed := false
	for key, value := range config.Paths {
		h, err := HashPath(key)
		if err != nil {
			fmt.Printf("Failed to sync: %s\n", err.Error())
			return
		}
		if h != value.Hash {
			changed = true
		}
	}
	_ = changed
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
	PATH_REPLACERS = GetPathReplacers()
	// get keys
	PATH_REPLACERS_SORTED = make([]string, len(PATH_REPLACERS))
	i := 0
	for k := range PATH_REPLACERS {
		PATH_REPLACERS_SORTED[i] = k
		i += 1
	}
	// sort keys by length in descending order
	slices.SortFunc(PATH_REPLACERS_SORTED, func(a string, b string) int {
		if len(a) > len(b) {
			return -1
		} else if len(a) == len(b) {
			return 0
		} else {
			return 1
		}
	})

	config, show_intro := get_config()
	defer save_config(&config)

	// checking git version
	git_version, err := run_cmd("git", "--version")
	if err != nil {
		fmt.Println("Convene uses *git* to synchronize your files. Please make sure it is installed and on your PATH.")
		return
	}
	fmt.Println("# Welcome to Convene!")

	fmt.Println("- Using", git_version)
	if show_intro || config.RepoLink == "" {
		onboarding(&config)
	}

	if len(os.Args) == 1 {
		print_usage()
		return
	}

	switch os.Args[1] {
	case "add":
		{
			cmd_add(&config, os.Args[2:])
		}
	case "rm":
		{
			cmd_rm(&config, os.Args[2:])
		}
	case "list":
		{
			cmd_list(&config)
		}
	case "sync":
		{
			cmd_sync(&config, os.Args[2:])
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
}
