package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cli/go-gh"
	"github.com/lspaccatrosi16/go-cli-tools/input"
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
	baseNames := map[string]bool{}

	for _, ent := range entries {
		ext := filepath.Ext(ent)
		_, fileName := filepath.Split(ent)

		withoutExt := fileName[:len(fileName)-len(ext)]

		_, parentFolderName := filepath.Split(filepath.Dir(ent))

		if parentFolderName == "out" {
			continue
		}

		newName := fmt.Sprintf("%s-%s%s", withoutExt, parentFolderName, ext)
		baseNames[newName] = true

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

	cont, err := input.GetConfirmSelection("Create a release using gh")
	if err != nil {
		panic(err)
	}

	if !cont {
		return
	}

	tag := input.GetValidatedInput("Git tag", func(in string) error {
		s := strings.Split(in, ".")
		if len(s) != 3 {
			return fmt.Errorf("git tag must have 3 components, not %d", len(s))
		}

		for i, n := range s {
			_, err := strconv.ParseInt(n, 10, 64)
			if err != nil {
				return fmt.Errorf("component %d is not an integer", i+1)
			}
		}
		return nil
	})

	createCommandText := fmt.Sprintf("release create v%s --generate-notes", tag)
	genAssetStr := ""

	for k := range baseNames {
		genAssetStr += fmt.Sprintf("\"%s\" ", k)
	}

	uploadReleaseText := fmt.Sprintf("release upload v%s %s", tag, genAssetStr)

	fmt.Println("Release commands:", "\n", "")
	fmt.Println("gh "+createCommandText, "\n", "")
	fmt.Println("gh " + uploadReleaseText)

	proceed, err := input.GetConfirmSelection("Enact release?")
	if err != nil {
		panic(err)
	}

	if !proceed {
		return
	}

	var stdout, stderr bytes.Buffer

	stdout, stderr, err = gh.Exec(strings.Split(createCommandText, " ")...)

	if err != nil {
		fmt.Println(stderr.String())
		panic(err)
	}

	fmt.Println(stdout.String())

	stdout, stderr, err = gh.Exec(strings.Split(uploadReleaseText, " ")...)

	if err != nil {
		fmt.Println(stderr.String())
		panic(err)
	}

	fmt.Println(stdout.String())

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
