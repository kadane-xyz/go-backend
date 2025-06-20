package responses

type StarredResponse struct {
	Id      any
	Starred bool
}

func NewStarredResponse(id any, starred bool) *StarredResponse {
	return &StarredResponse{
		Id:      id,
		Starred: starred,
	}
}
