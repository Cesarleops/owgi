package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: owgie <command> [<args>...]\n")
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

	case "hash-object":
		param := os.Args[3]

		file, err := os.ReadFile(param)
		if err != nil {
			fmt.Println("unable to find file")
			return
		}
		h := sha1.New()
		hashInput := "blob " + strconv.Itoa(len(file)) + "\x00" + string(file)
		// fmt.Println("hash input", hashInput)

		h.Write([]byte(hashInput))

		hash := fmt.Sprintf("%x", h.Sum(nil))

		path := ".git/objects/" + hash[0:2] + "/" + hash[2:]

		errCreating := os.MkdirAll(".git/objects/"+hash[0:2], 0775)

		if errCreating != nil {
			fmt.Println("fuck 1")
			return
		}

		f, err := os.Create(path)

		if err != nil {
			fmt.Println("fuck 2")
			return
		}
		zlibW := zlib.NewWriter(f)
		_, errCom := zlibW.Write([]byte(hashInput))
		if errCom != nil {
			fmt.Println("fuck 3")
			return
		}
		zlibW.Close()
		fmt.Println(hash)

	case "cat-file":

		param := os.Args[3]

		file, err := os.ReadFile(".git/objects/" + param[0:2] + "/" + param[2:])

		if err != nil {
			fmt.Println("Im a retard")
			return
		}

		byteReader := bytes.NewReader(file)

		zlibReader, err := zlib.NewReader(byteReader)

		if err != nil {
			fmt.Printf("Something went wrong decompressing %v\n", err)
			os.Exit(1)
		}

		//Don't forget to close the reader and avoid memory leaks
		defer zlibReader.Close()

		//decompressed bytes
		decompressedData, err := io.ReadAll(zlibReader)

		if err != nil {
			fmt.Printf("Error: Failed during decompression. %v\n", err)
			os.Exit(1)
		}

		//index of the null byte
		i := slices.Index(decompressedData, '\x00')

		content := decompressedData[i+1:]

		fmt.Printf("%s", string(content))

	default:
		os.Exit(1)

	}
}
