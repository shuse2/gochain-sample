package gochain

import (
	"crypto/rand"
	"fmt"
)

func PseudoUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return ""
	}
	return fmt.Sprintf("%X%X%X%X%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type StringSet struct {
	set map[string]bool
}

func NewStringSet() StringSet {
	return StringSet{make(map[string]bool)}
}

func (set *StringSet) Add(str string) bool {
	_, ok := set.set[str]
	set.set[str] = true
	return !ok
}

func (set *StringSet) Keys() []string {
	keys := []string{}
	for i, _ := range set.set {
		keys = append(keys, i)
	}
	return keys
}
