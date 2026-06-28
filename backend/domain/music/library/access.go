package library

type Visibility struct {
	s string
}

var (
	VisibilityPublic  = Visibility{s: "public"}
	VisibilityPrivate = Visibility{s: "private"}
)

func (v Visibility) String() string {
	return v.s
}

type AccessContext interface {
	CanSee(v Visibility) bool
}
