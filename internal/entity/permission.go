package entity

type Permission struct {
	ID              int64
	Code            string
	GrantedForUsers []User
}

// type UserPermission struct {
// 	UserId    int
// 	PermId    int
// 	GrantedAt time.Time
// }
