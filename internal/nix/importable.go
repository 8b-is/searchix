package nix

type Importable interface {
	BleveType() string
	GetName() string
	GetSource() string
}

func GetKey(i Importable) string {
	return i.BleveType() + "/" + i.GetSource() + "/" + i.GetName()
}
