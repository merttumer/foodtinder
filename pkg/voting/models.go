package voting

type Vote struct {
	ProductID string `json:"id"`
	SessionID string `json:"session_id"`
	Score     int    `json:"score"`
}

