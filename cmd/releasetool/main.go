package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	_, lastDir := filepath.Split(wd)

	if lastDir != "out" {
		fmt.Println("current directory is not out")
		os.Exit(1)
	}

	entries := crawlFolder(wd)

	summary := []string{}

	for _, ent := range entries {
		ext := filepath.Ext(ent)
		_, fileName := filepath.Split(ent)

		withoutExt := fileName[:len(fileName)-len(ext)]

		_, parentFolderName := filepath.Split(filepath.Dir(ent))

		if parentFolderName == "out" {
			continue
		}

		newName := fmt.Sprintf("%s-%s%s", withoutExt, parentFolderName, ext)

		src, err := os.Open(ent)

		if err != nil {
			panic(err)
		}

		dstLoc := filepath.Join(wd, newName)

		dst, err := os.Create(dstLoc)

		if err != nil {
			panic(err)
		}

		io.Copy(dst, src)

		src.Close()
		dst.Close()

		err = os.Chmod(dstLoc, 0o755)
		if err != nil {
			panic(err)
		}

		summary = append(summary, newName)
	}

	fmt.Println("Prepared Release Assets:")

	for _, a := range summary {
		fmt.Printf("%-40s OK\n", a)
	}
}

func crawlFolder(path string) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	found := []string{}

	for _, ent := range entries {
		if ent.IsDir() {
			sub := crawlFolder(filepath.Join(path, ent.Name()))
			found = append(found, sub...)
		} else {
			found = append(found, filepath.Join(path, ent.Name()))
		}

	}

	return found
}
