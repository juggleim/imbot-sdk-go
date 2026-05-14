package models

type UserInfo struct {
	UserId       string
	UserName     string
	UserPortrait string
	Extras       map[string]string
	UpdatedTime  int64
}
