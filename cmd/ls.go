package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/liclac/planetarium/warc"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists the index of a WARC file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
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
			switch hdr.WARCType {
			case warc.WARCTypeWARCInfo:
				fmt.Println("? warcinfo", hdr.ContentLength)
			case warc.WARCTypeRequest:
				fmt.Println("> request ", hdr.ContentLength, hdr.Fields.Get("WARC-Target-URI"))
			case warc.WARCTypeResponse:
				fmt.Println("< response", hdr.ContentLength, hdr.Fields.Get("WARC-Target-URI"))
			case warc.WARCTypeMetadata:
				fmt.Println("& metadata", hdr.ContentLength, hdr.Fields.Get("WARC-Target-URI"))
			case warc.WARCTypeResource:
				fmt.Println("% resource", hdr.ContentLength, hdr.Fields.Get("WARC-Target-URI"))
			default:
				fmt.Fprintln(os.Stderr, "- unknown record type: ", hdr.WARCType)
				return nil
			}

			switch hdr.WARCType {
			case warc.WARCTypeWARCInfo,
				warc.WARCTypeMetadata:
				if _, err := io.Copy(os.Stderr, block); err != nil {
					return err
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
