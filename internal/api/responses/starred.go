package responses

type StarredResponse struct {
	Id      int32
	Starred bool
}

func NewStarredResponse(id int32, starred bool) *StarredResponse {
	return &StarredResponse{
		Id:      id,
		Starred: starred,
	}
}
