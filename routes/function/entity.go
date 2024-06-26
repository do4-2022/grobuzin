package function

type CreateFunctionDTO struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description" binding:"required"`
	Language    string            `json:"language" binding:"required"`
	Files       map[string]string `json:"files" binding:"required"`
}

type GetFunctionDTO struct {
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Language       string            `json:"language"`
	Files          map[string]string `json:"files"`
	Built          bool              `json:"built"`
	BuildTimestamp int64             `json:"build_timestamp"`
	OwnerID        int               `json:"owner_id"`
}

type ExecutionResultDTO struct {
	Time     int64             `json:"time"`
	Response ExecutionResponse `json:"response"`
}
type ExecutionResponse struct {
	Status  int               `json:"status"`
	Body    interface{}       `json:"body"`
	Headers map[string]string `json:"headers"`
}
