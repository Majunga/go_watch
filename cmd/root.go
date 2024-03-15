package cmd

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yargevad/filepathx"
)

var verbose *bool

var rootCmd = &cobra.Command{
	Use:   "go_watch pattern command\n\n \tpattern: standard glob wildcards such as **/*\n \tcommand: is a single string command, chain commands by using &&\n\n",
	Short: "File watcher that excutes a command when something changes",
	Long:  ``,
	Args:  cobra.ExactArgs(2),
	Run:   RootCommandHandler,
}

func RootCommandHandler(cmd *cobra.Command, args []string) {
	pattern := args[0]
	command := args[1]
	log.Printf("Watching %s", pattern)
	watch(func() []file { return files_to_watch(pattern) }, command)

	<-make(chan struct{})
}

func files_to_watch(pattern string) []file {
	matches, err := filepathx.Glob(pattern)

	if err != nil {
		log.Fatalf("Invalid Pattern: %s", err)
	}

	return map_paths(matches)
}

func watch(pathsToWatch func() []file, command string) {
	paths := pathsToWatch()
	if *verbose {
		log.Printf("Files being watched: %s", paths)
	}

	go func() {
		for {
			<-time.After(100 * time.Millisecond)
			for _, watch := range paths {
				hash, err := fileHash(watch.path)
				if err != nil {
					log.Fatal(err)
				}

				hashString := *hash
				if hashString != watch.hash {
					log.Printf("Path has been modified: %s", watch.path)
					executeCommand(command)

					paths = pathsToWatch()
					break
				}

				watch.hash = *hash
			}
		}
	}()
}

func executeCommand(fullCommand string) {
	commands := strings.Split(fullCommand, "&&")

	for _, command := range commands {
		commandParts := trim(command)

		var cmd *exec.Cmd
		if len(commandParts) > 1 {
			args := commandParts[1:]
			cmd = exec.Command(commandParts[0], args...)
		} else {
			cmd = exec.Command(commandParts[0])
		}

		stdout, err := cmd.Output()

		if err != nil {
			stdErr := err.(*exec.ExitError)
			log.Printf("%s", stdErr.Stderr)
		}

		log.Printf("%s", stdout)
	}
}

func trim(commandString string) []string {
	result := []string{}
	for _, s := range strings.Split(strings.TrimSpace(commandString), " ") {
		result = append(result, strings.TrimSpace(s))
	}

	return result
}

type file struct {
	path string
	hash string
}

func map_paths(s []string) []file {
	mapped := []file{}

	for _, a := range s {
		stat, err := os.Stat(a)
		if err != nil {
			log.Fatal(err)
		}

		if stat.IsDir() {
			if *verbose {
				log.Printf("Ignoring directory %s", a)
			}
			continue
		}

		hash, err := fileHash(a)
		if err != nil {
			log.Fatal(err)
		}

		mapped = append(mapped, file{path: a, hash: *hash})
	}

	return mapped
}

func fileHash(path string) (*string, error) {
	osFile, err := os.Open(path)
	if err != nil && os.IsExist(err) {
		return nil, err
	}

	defer osFile.Close()

	h := sha256.New()

	if _, err := io.Copy(h, osFile); err != nil {
		return nil, err
	}

	hash := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return &hash, nil
}

func Init() {
	verbose = rootCmd.Flags().BoolP("verbose", "v", false, "Used to print more verbosely to find issues")
}

func Execute() {
	Init()
	log.Println(verbose)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
