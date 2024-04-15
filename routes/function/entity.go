package function

type FunctionDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Language    string `json:"language" binding:"required"`
	Code        string `json:"code" binding:"required"`
}
