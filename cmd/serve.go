package cmd

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve <amber>",
	Short: "Serves a mirror from an 'import'",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cacheDir := args[0]

		// Make sure the amber cache exists, is a directory, and isn't empty.
		cacheEntries, err := ioutil.ReadDir(cacheDir)
		if err != nil {
			return err
		}
		if len(cacheEntries) == 0 {
			return fmt.Errorf("no files preserved in amber; remember to run 'planetarium import'!")
		}

		mirrorDomains := map[string]string{}
		for _, domainDef := range viper.GetStringSlice("domain") {
			parts := strings.SplitN(domainDef, "=", 2)
			if len(parts) < 2 {
				return fmt.Errorf("invalid domain definition: %s", domainDef)
			}
			mirrorDomains[parts[0]] = parts[1]
		}

		lis, err := net.Listen("tcp", viper.GetString("addr"))
		if err != nil {
			return err
		}
		fmt.Println("Listening on", lis.Addr().String())

		return http.Serve(lis, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			reqURL := *req.URL
			if reqURL.Host == "" {
				reqURL.Host = req.Host
			}
			if realDomain, ok := mirrorDomains[reqURL.Hostname()]; ok {
				reqURL.Host = realDomain
			}
			u := normaliseURL(&reqURL)

			if err := serveFromCache(rw, u, cacheDir); err != nil {
				if os.IsNotExist(err) {
					rw.WriteHeader(http.StatusNotFound)
					fmt.Fprintln(rw, "Not Found")
					return
				}
				rw.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(rw, "ERROR:", err)
				return
			}
		}))
	},
}

func serveFromCache(rw http.ResponseWriter, u *url.URL, cacheDir string) error {
	filename := normURLToPath(cacheDir, u)

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	rsp, err := http.ReadResponse(bufio.NewReader(f), nil)
	if err != nil {
		return err
	}
	if rsp.Body != nil {
		defer rsp.Body.Close()
	}

	for k, vs := range rsp.Header {
		rw.Header()[k] = vs
	}
	rw.WriteHeader(rsp.StatusCode)
	io.Copy(rw, rsp.Body)
	return nil
}

func normaliseURL(u *url.URL) *url.URL {
	return &url.URL{
		Scheme:   "http",
		Host:     u.Host,
		Path:     u.Path,
		RawPath:  u.RawPath,
		RawQuery: u.Query().Encode(),
	}
}

func normURLToPath(base string, uu *url.URL) string {
	u := *uu
	u.Scheme = ""
	return filepath.Join(base, strings.TrimRight(u.String(), "/")+".http")
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringP("addr", "a", "127.0.0.1:8000", "address to listen on")
	serveCmd.Flags().StringSliceP("domain", "d", nil, "my-example-mirror.com=example.com")
	viper.BindPFlags(serveCmd.Flags())
}
