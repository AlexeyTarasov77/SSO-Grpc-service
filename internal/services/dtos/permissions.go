package dtos

type GetPermissionOptionsDTO struct {
	Code string
	ID   int64
}

type FetchManyPermissionsOptionsDTO struct {
	Ids   []int
	Codes []string
}
