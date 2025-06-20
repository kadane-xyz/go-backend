package requests

type RequestQueryParamOrder string

const (
	RequestQueryParamOrderASC  RequestQueryParamOrder = "asc"
	RequestQueryParamOrderDesc RequestQueryParamOrder = "desc"
)

func (p RequestQueryParamOrder) IsValid() bool {
	switch p {
	case RequestQueryParamOrderASC, RequestQueryParamOrderDesc:
		return true
	default:
		return false
	}
}

func (p RequestQueryParamOrder) String() string {
	return string(p)
}
