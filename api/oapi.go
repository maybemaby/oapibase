package api

import "github.com/oaswrap/spec/option"

func Responses(responses map[int]any) option.OperationOption {

	return func(oc *option.OperationConfig) {
		for code, schema := range responses {
			option.Response(code, schema)(oc)
		}
	}

}

func ResponsesWithDefault(responses map[int]any) option.OperationOption {
	return func(oc *option.OperationConfig) {
		option.Response(500, new(ServerErrorResponse))(oc)
		for code, schema := range responses {
			option.Response(code, schema)(oc)
		}
	}
}

type ServerErrorResponse struct {
	Message string `json:"message" example:"Internal Server Error" required:"true"`
	Status  int    `json:"status" enum:"500" required:"true"`
}

func DefaultServerErrorResponse() ServerErrorResponse {
	return ServerErrorResponse{
		Message: "Internal Server Error",
		Status:  500,
	}
}

type AuthErrorResponse struct {
	Message string `json:"message" example:"Unauthorized" required:"true"`
	Status  int    `json:"status" enum:"401" required:"true"`
}

func DefaultAuthErrorResponse() AuthErrorResponse {
	return AuthErrorResponse{
		Message: "Unauthorized",
		Status:  401,
	}
}
