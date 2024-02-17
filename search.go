package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"

	"slices"

	"github.com/inconshreveable/log15"
	"github.com/joomcode/errorx"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var ParseError = errorx.NewType(errorx.CommonErrors, "parse_error")
var ReadError = errorx.NewType(errorx.CommonErrors, "read_error")

// certExtensions defines the valid certificate file extensions.
var certExtensions = []string{".crt", ".pem", ".cer"}

// CertEntry represents a certificate with metadata
type CertEntry struct {
	// Path represents where on the filesystem the certificate was found
	Path string `json:"path"`
	// Raw is the raw certificate data
	Raw []byte `json:"raw"`
	// Index is the index of the certificate in the file
	Index int `json:"index"`
}

// searchCmd represents the search command.
var searchCmd = &cobra.Command{
	Use:   "search [path]",
	Short: "Search recursively under a given path for certificates",
	Long: `
Search recursively under a given path for certificates. 
It saves the certificates it finds in a single file under /tmp/.cert-inspector/cache`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := log15.New(log15.Ctx{"module": "search"})
		logger.Info("Searching under path", "path", args[0])
		fs := afero.NewOsFs()
		// Convert the path to an absolute path
		path, _ := filepath.Abs(args[0])
		certs, err := SearchAndParse(fs, path)
		if err != nil {
			logger.Error("Failed to search and parse", "error", err)
			os.Exit(1)
		}
		// Write the certificates found to a cache file
		logger.Info("Found certificates", "count", len(certs))
		file, err := fs.OpenFile("/tmp/.cert-inspector/cache", os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Error("Failed to open cache file", "error", err)
			os.Exit(1)
		}
		for _, cert := range certs {
			contents, _ := json.MarshalIndent(cert, "", "    ")
			_, err = file.Write(contents)
			if err != nil {
				logger.Error("Failed to write certificate to cache", "path", cert.Path, "error", err)
			}
		}
	},
}

// SearchAndParse recursively searches for certificate files under the given path.
// It returns a slice of x509.Certificate pointers and an error, if any.
func SearchAndParse(fs afero.Fs, path string) ([]CertEntry, error) {
	var certs []CertEntry
	err := afero.Walk(fs, path, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		// Check if it's a cert file using the extension
		if slices.Contains(certExtensions, filepath.Ext(info.Name())) && info.Size() > 0 {
			// Try to read and parse it. If it's a valid cert, add it to the list
			// NOTE: There can be multiple certs in a single file. e.g. a trust bundle
			contents, err := afero.ReadFile(fs, p)
			if err != nil {
				return ReadError.Wrap(err, "failed to read file at %s", p)
			}
			var block *pem.Block
			idx := 0
			for len(contents) > 0 {
				block, contents = pem.Decode(contents)
				if block == nil {
					break
				}
				if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
					return ParseError.New("invalid certificate")
				}
				certBytes := block.Bytes
				cert, err := x509.ParseCertificate(certBytes)
				if err != nil {
					return ParseError.Wrap(err, "failed to parse certificate at %s, index %d", p, idx)
				}
				certs = append(certs, CertEntry{Path: p, Raw: cert.Raw, Index: idx})
				idx++
			}
		}
		return nil
	})
	return certs, err
}
