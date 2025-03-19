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
$> cert-inspector search /etc/
$> cert-inspector tree
```