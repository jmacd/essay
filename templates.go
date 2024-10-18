package essay

import (
	"path"
)

// var (
// 	tmplPath = path.Join(
// 		strings.Split(os.Getenv("GOPATH"), ";")[0],
// 		"src/github.com/lightstep/sandbox/jmacd/essay/tmpl",
// 	)
// )

const tmplPath = "/Users/josh.macdonald/src/essay/tmpl" // @@@

func templateGlobPath() string {
	return path.Join(tmplPath, "*")
}
