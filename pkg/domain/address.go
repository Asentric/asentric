package domain

type Address string

func (a Address) String() string {
	return string(a)
}

func (a Address) Hex() string {
	return a.String()
}

func (a Address) IsZero() bool {
	return a.String() == "" || a.String() == "0x0000000000000000000000000000000000000000"
}
