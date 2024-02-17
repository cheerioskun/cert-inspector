package main

import "crypto/x509"

// CertEntry represents a certificate with metadata
type CertEntry struct {
	// Path represents where on the filesystem the certificate was found
	Path string `json:"path"`
	// Raw is the raw certificate data
	Raw []byte `json:"raw"`
	// Index is the index of the certificate in the file
	Index int `json:"index"`
	// Certificate is the parsed x509 certificate, not saved to JSON
	Certificate *x509.Certificate `json:"-"`
}
