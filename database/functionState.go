package database

type FnStatusCode int

const (
	FnCreating FnStatusCode = iota
	FnReady
	FnRunning
	FnUnknownState
)

// This struct represents an instance of a function, especially it's address and status 
// Each key in the redis is namespaced as it follows:
// <id of the function>:<id of the instance>
type FunctionState struct {
	ID			string			`redis:"address"`
	Address 	string			`redis:"address"`
	Port    	uint16			`redis:"port"`
	Status 		FnStatusCode 	`redis:"status"`
	LastUsed	string			`redis:"lastUsed"`
}