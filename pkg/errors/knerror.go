package errors

func NewKNError(msg string) *KNError {
	return &KNError{
		msg: msg,
	}
}

func (kne *KNError) Error() string {
	return kne.msg
}
