package routing

import "strings"

type List []Record

func NewList() List {
	return List{}
}

func (List) TableName() string {
	return "routings"
}

func (List) Permissions(p ...string) string {
	return ""
}

func (List) Owner(o ...uint) uint {
	return 0
}

func (List) Users(u ...uint) []uint {
	return []uint{}
}

func (List) Groups(g ...uint) []uint {
	return []uint{}
}

func (List) IDValue(...uint) uint {
	return 0
}

func (List) UniqueCode(uc ...string) string {
	return ""
}

func (List) Complete() error {
	return nil
}

func (List) Hasher() error {
	return nil
}

func (List) Prepare() error {
	return nil
}
func (e List) Len() int {
	return len(e)
}
func (e List) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
func (e List) Less(i, j int) bool {
	return pathLength(e[i].Path) > pathLength(e[j].Path)
}
func pathLength(p string) uint {
	ts := strings.Split(p, "/")
	l := uint(len(ts))
	if ts[len(ts)-1] != "" {
		l++
	}
	return l
}
