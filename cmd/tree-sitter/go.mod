module github.com/alecthomas/chroma/v2/cmd/chroma

go 1.22

toolchain go1.24.5

replace github.com/alecthomas/chroma/v2 => ../../

require (
	github.com/alecthomas/chroma/v2 v2.19.0
	github.com/alecthomas/kong v1.12.1
	github.com/mattn/go-colorable v0.1.14
	github.com/mattn/go-isatty v0.0.20
)

require (
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/nushell/tree-sitter-nu v0.0.0-20250716021349-6544c4383643 // indirect
	github.com/smacker/go-tree-sitter v0.0.0-20240827094217-dd81d9e9be82 // indirect
	golang.org/x/sys v0.29.0 // indirect
)
