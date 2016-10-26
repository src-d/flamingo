package flamingo

import "fmt"

var (
	erroMsgNotAllowedStr     = "Message tried to send is not allowed for broadcasting"
	erroSomeMsgLostFormatter = "%d messages could not be delivered"
	erroAllMsgLostFormatter  = "No messages sent; %d messages lost"
)

type ErrorNotValidMessage struct{ error }

type ErrorNoMessageSent struct{ error }

type ErrorSomeMessagesFailed struct{ error }

func NewErrorNotValidMessage() ErrorNotValidMessage {
	return ErrorNotValidMessage{fmt.Errorf(erroMsgNotAllowedStr)}
}

func NewErrorNoMessageSent(v ...interface{}) ErrorNoMessageSent {
	return ErrorNoMessageSent{fmt.Errorf(erroAllMsgLostFormatter, v...)}
}

func NewErrorSomeMessagesFailed(v ...interface{}) ErrorSomeMessagesFailed {
	return ErrorSomeMessagesFailed{fmt.Errorf(erroSomeMsgLostFormatter, v...)}
}
