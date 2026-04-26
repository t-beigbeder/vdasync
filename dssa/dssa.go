package dssa

type Path []string

type Rights struct {
	Read    bool
	Write   bool
	Execute bool
}

type DataEntry struct {
	IsDir         bool
	Path          []string
	Size          int64
	Mtime         int64
	User          int
	UserRights    Rights
	Group         int
	GroupRights   Rights
	OtherRights   Rights
	IsSymLink     bool
	SymLinkTarget string
	Error         error
}

type Dssa interface {
	List(Path) ([]*DataEntry, error)
	Stat(Path) (*DataEntry, error)
	SetStat(*DataEntry) error
}
