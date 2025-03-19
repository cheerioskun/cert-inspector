# Cert Inspector

Cert Inspector is a command-line tool designed to inspect and visualize all certificates on a system.

## Features

- Collect all certificates from a filesystem recursively
- Display certificate information concisely with their relationships

## Future Additions
- Identify and highlight expired or invalid certificates
- Export certificate data to various formats (e.g., JSON, CSV)

## Installation
It can be installed using 
```sh
go install github.com/cheerioskun/cert-inspector@latest
```

Or you can build it from source like so
```sh
go build -o cert-inspector ./...
```

## Usage
The tool has two commands: search and tree. Search is used to crawl the filesystem and index all the certs. Tree is used to visualize the temporary index created.

```sh
> cert-inspector search ./testdata
INFO[03-19|16:21:09] Searching under path                     module=search path=./testdata
INFO[03-19|16:21:09] Found certificates                       module=search count=23

> cert-inspector tree
INFO[03-19|16:21:09] Loaded certificates                      module=tree count=23
DBUG[03-19|16:21:09] Created certificate forest               module=tree trees=2
Example Root CA(2-1):testdata/certs/chain/leaf/example-service.chain.pem
└── Example Intermediate CA(2-1):testdata/certs/chain/intermediate/intermediate-ca.crt.pem
    └── example-service.example.com(2-0):testdata/certs/chain/leaf/example-service.chain.pem
duplicate-cert.example.com(3-0):testdata/certs/duplicates/duplicate-cert.crt
```

The output reference is 
```
<CommonName>(<FREQ>-<CHILD>):<PATH TO FIRST INSTANCE>

Example Root CA(2-1):testdata/certs/chain/leaf/example-service.chain.pem

CommonName = Example Root CA
Frequency = 2 (it was seen two times)
Children = 1 (only a single certificate has this issuer)
Path = testdata/certs/chain/leaf/example-service.chain.pem
```

