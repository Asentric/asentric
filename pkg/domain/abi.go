package domain

// ABIRegistry provides semantic access to ABI metadata
// This is the semantic backbone for rule evaluation
type ABIRegistry interface {
	GetMethod(address Address, selector string) (Method, bool)
	GetEvent(address Address, topic Hash) (Event, bool)
}

// Method represents decoded method metadata
type Method struct {
	Name string
	Args []ABIArg
}

// ABIArg represents a method argument
type ABIArg struct {
	Name string
	Type string
}
