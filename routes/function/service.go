package function

var functions = make(map[string]Function)

func GetAllFunctions() []Function {
	var array []Function
	for _, value := range functions {
		array = append(array, value)
	}
	return array
}