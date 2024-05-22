package database

type FnStatusCode int

const ( 
	FnReady 		FnStatusCode = iota
	FnRunning
	FnUnknownState
)

// This struct represents an instance of a function, especially it's address and status 
// Each key in the redis is namespaced as it follows:
// <id of the function>:<id of the instance>
type FunctionState struct {
	ID			string	`redis:"id"`
	Address 	string	`redis:"address"`
	Port    	uint16	`redis:"port"`
	Status 		int		`redis:"status"`
	LastUsed	string	`redis:"lastUsed"`
}