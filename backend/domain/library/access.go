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

type LoggedInAccessContext struct{}

func NewLoggedInAccessContext() LoggedInAccessContext {
	return LoggedInAccessContext{}
}

func (LoggedInAccessContext) CanSee(Visibility) bool {
	return true
}

type AnonymousAccessContext struct{}

func NewAnonymousAccessContext() AnonymousAccessContext {
	return AnonymousAccessContext{}
}

func (AnonymousAccessContext) CanSee(v Visibility) bool {
	return v.Public()
}
