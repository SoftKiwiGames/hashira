package hashira

import "github.com/qbart/hashira/ds"

type World struct {
	Maps ds.HashMap[string, *Map]
}
