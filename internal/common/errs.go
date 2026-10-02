package common

type errorConst string

const ErrUnhandledFileType errorConst = "unhandled file type"
const ErrReadClosedQueue errorConst = "all is read on closed queue"
const ErrNeededRmForbidden errorConst = "needed removal forbidden"
const ErrNeededWriteEnableForbidden errorConst = "needed write enablement forbidden"

func (e errorConst) Error() string {
	return string(e)
}
