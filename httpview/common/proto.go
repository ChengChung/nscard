package common

type ErrorMessage struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

func HandleError(err error) ErrorMessage {
	if err == nil {
		return ErrorMessage{Ok: true, Message: "success"}
	}
	return ErrorMessage{Ok: false, Message: err.Error()}
}

func HandleSuccess(message string) ErrorMessage {
	return ErrorMessage{Ok: true, Message: message}
}
