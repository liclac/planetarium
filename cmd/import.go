package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/liclac/planetarium/warc"
)

var importCmd = &cobra.Command{
	Use:   "import <amber> <file1.warc> [fileN.warc...]",
	Short: "Adds a WARC file to a piece of amber, for 'serve'",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cachedir := args[0]
		filenames := args[1:]

		for _, filename := range filenames {
			if err := importFile(cachedir, filename); err != nil {
				return err
			}
		}

		return nil
	},
}

func importFile(cacheDir, filename string) error {
	fmt.Println("Importing file...", filename)

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	r := warc.NewReader(f)
	for {
		hdr, block, err := r.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if hdr.WARCType != warc.WARCTypeResponse {
			fmt.Println(hdr.WARCType)
			continue
		}

		rawURL := strings.Trim(hdr.Fields.Get("WARC-Target-URI"), "<>")
		if rawURL == "" {
			fmt.Fprintln(os.Stderr, "response is missing WARC-Target-URI")
		}
		fmt.Println(hdr.WARCType, rawURL)

		u, err := url.Parse(rawURL)
		if err != nil {
			return err
		}
		u = normaliseURL(u)

		path := normURLToPath(cacheDir, u)
		if err := copyToFile(path, block); err != nil {
			return err
		}
	}

	return nil
}

func copyToFile(filename string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(importCmd)
}
