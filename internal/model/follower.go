package model

type FollowAction struct {
	TargetUserId uint32 `json:"target_user_id"`
	SenderUserId uint32 `json:"sender_user_id"`
	CreatedAt    string `json:"created_at"`
}
