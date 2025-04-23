package entity

type Permission struct {
	ID   int64
	Code string
}

type Permissions []*Permission

func (self Permissions) Includes(code string) (includes bool) {
	for _, perm := range self {
		if perm.Code == code {
			includes = true
			break
		}
	}
	return
}
