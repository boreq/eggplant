package library

var defaultVisibility = NewVisibility(false)

type Visibility struct {
	public bool
}

func NewVisibility(public bool) Visibility {
	return Visibility{public: public}
}

func (v Visibility) Public() bool {
	return v.public
}

type AccessContext interface {
	CanSee(v Visibility) bool
}
