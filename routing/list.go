package routing

type List map[string]Record

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
