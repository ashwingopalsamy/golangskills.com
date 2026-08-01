package apicompatibility

type Request struct {
	UserId string
}

func Subject(request Request) string { return request.UserId }
