package universe

import (
	"github.com/jmacd/essay"
)

func (u *Universe) String() string {
	return u.name
}

func (u *Universe) Display(doc essay.Document) {
	doc.Section("Seed", u.seed)
	doc.Section("Etc", "...")
}
