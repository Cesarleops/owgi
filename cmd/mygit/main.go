package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type TreeEntry struct {
	shal     []byte
	filemode string
	filename string
}

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: owgie <command> [<args>...]\n")
		os.Exit(1)
	}

	// fileModes := map[string]int{
	// 	"normal":    100644,
	// 	"exe":       100755,
	// 	"symbolic":  120000,
	// 	"directory": 400000,
	// }
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

		filePath := os.Args[3]
		// fmt.Println("ff", filePath)
		file, err := os.ReadFile(filePath)

		if err != nil {
			fmt.Println("err reading", err)
			os.Exit(1)
		}

		h := sha1.New()

		hashInput := "blob " + strconv.Itoa(len(file)) + "\x00" + string(file)

		h.Write([]byte(hashInput))

		hash := fmt.Sprintf("%x", h.Sum(nil))

		path := ".git/objects/" + hash[0:2] + "/" + hash[2:]

		errCreating := os.MkdirAll(".git/objects/"+hash[0:2], 0775)

		if errCreating != nil {
			fmt.Println("err c", errCreating)
			os.Exit(1)
		}

		f, err := os.Create(path)

		if err != nil {
			fmt.Println("err c2", err)
			os.Exit(1)
		}

		zlibW := zlib.NewWriter(f)

		_, errCom := zlibW.Write([]byte(hashInput))

		if errCom != nil {
			os.Exit(1)
		}

		zlibW.Close()

		fmt.Println(hash)

	case "cat-file":

		shal := os.Args[3]

		file, err := os.ReadFile(".git/objects/" + shal[0:2] + "/" + shal[2:])

		if err != nil {
			os.Exit(1)
			return
		}

		byteReader := bytes.NewReader(file)
		zlibReader, err := zlib.NewReader(byteReader)

		defer zlibReader.Close()

		if err != nil {
			fmt.Printf("Something went wrong decompressing %v\n", err)
			os.Exit(1)
		}

		//close the reader and avoid memory leaks

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

	case "ls-tree":

		// fmt.Println("args", os.Args)

		flag := os.Args[2]
		treeHash := os.Args[3]

		treePath := ".git/objects/" + treeHash[0:2] + "/" + treeHash[2:]

		// fmt.Println("tree", treeHash)

		file, err := os.ReadFile(treePath)

		if err != nil {
			fmt.Println("ERR:", err)
			os.Exit(1)
		}

		byteReader := bytes.NewReader(file)

		zlibReader, err := zlib.NewReader(byteReader)

		data, err := io.ReadAll(zlibReader)

		i := slices.Index(data, '\x00')

		files := ParseFiles(data[i+1:])
		sort.Slice(files, func(i, j int) bool {
			return files[i].filename < files[j].filename
		})

		if flag == "--name-only" {
			for _, f := range files {
				fmt.Println(f.filename)
			}
		}

		defer zlibReader.Close()

		if err != nil {
			fmt.Printf("Something went wrong decompressing %v\n", err)
			os.Exit(1)
		}

	default:
		os.Exit(1)

	}
}

func ParseFiles(files []byte) []TreeEntry {

	//Here i must extract the file mode, name and hash from each file, all of them are separated
	// with a null byte

	entries := make([]TreeEntry, 0)

	i := 0
	entry := ""

	for i < len(files) {

		//Everything before first null byte is the filemode and filename,then
		// after the null byte, the 20 next characters are the hash and thats it

		if files[i] == '\x00' {
			start := i + 1
			end := start + 20
			hash := files[start:end]
			// fmt.Printf("%s %x \n", entry, hash)
			splittedEntry := strings.Split(entry, " ")
			// fmt.Printf("%q \n", splittedEntry[1])
			tEntry := TreeEntry{
				filemode: splittedEntry[0],
				filename: splittedEntry[1],
				shal:     hash,
			}
			entries = append(entries, tEntry)
			i = end
			entry = ""
			continue
		}
		entry += string(files[i])

		i++

	}
	return entries
}
