package main

import "github.com/joomcode/errorx"

var (
	ParseError = errorx.NewType(errorx.CommonErrors, "parse_error")
	ReadError  = errorx.NewType(errorx.CommonErrors, "read_error")
)
