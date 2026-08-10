package common

type errorConst string

const ErrUnhandledFileType errorConst = "unhandled file type"

func (e errorConst) Error() string {
	return string(e)
}
