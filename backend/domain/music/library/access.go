package library

type Visibility struct {
	s string
}

var (
	VisibilityPublic  = Visibility{"public"}
	VisibilityPrivate = Visibility{"private"}
)

func NewVisibility(public bool) Visibility {
	if public {
		return VisibilityPublic
	}
	return VisibilityPrivate
}

func (v Visibility) Public() bool {
	return v == VisibilityPublic
}

type AccessContext interface {
	CanSee(v Visibility) bool
}
