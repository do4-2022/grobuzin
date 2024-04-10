package function

type FunctionDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Language    string `json:"language" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

type Function struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Code        string `json:"code"`
}

func dtoToFunction(id string, dto FunctionDTO) Function {
	return Function{
		ID:          id,
		Name:        dto.Name,
		Description: dto.Description,
		Language:    dto.Language,
		Code:        dto.Code,
	}
}
