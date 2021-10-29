package essay

import (
	"os"
	"path"
	"strings"
)

var (
	tmplPath = path.Join(
		strings.Split(os.Getenv("GOPATH"), ";")[0],
		"src/github.com/lightstep/sandbox/jmacd/essay/tmpl",
	)
)

func templateGlobPath() string {
	return path.Join(tmplPath, "*")
}
