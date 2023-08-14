package session

type UserSession struct {
	SessionID string `json:"session_id"`
	ExpireAt  int64  `json:"expire_at"`
}
