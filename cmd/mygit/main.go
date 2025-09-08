package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"slices"
)

// Usage: your_program.sh <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")
	case "cat-file":
		param := os.Args[3]
		file, err := os.ReadFile(".git/objects/" + param[0:2] + "/" + param[2:])
		if err != nil {
			fmt.Println("dumb")
			return
		}
		byteReader := bytes.NewReader(file)
		zlibReader, err := zlib.NewReader(byteReader)
		if err != nil {
			fmt.Printf("Error: Failed to create zlib reader. The object might be corrupt. %v\n", err)
			os.Exit(1)
		}
		defer zlibReader.Close()
		decompressedData, err := io.ReadAll(zlibReader)
		if err != nil {
			fmt.Printf("Error: Failed during decompression. %v\n", err)
			os.Exit(1)
		}
		i := slices.Index(decompressedData, '\x00')
		content := decompressedData[i+1:]
		fmt.Printf("%s", string(content))

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
