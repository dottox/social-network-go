package model

type Role struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Level       int    `json:"level"` // 0 = user, 1 = mod, 2 = admin
	Description string `json:"description"`
}
