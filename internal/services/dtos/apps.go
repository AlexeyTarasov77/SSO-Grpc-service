package dtos

type GetOrCreateAppDTO struct {
	AppID     int64
	IsCreated bool
}

type GetAppOptionsDTO struct {
	AppID   int32
	AppName string
}
