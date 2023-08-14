package voting

type AvgVoteResponse struct {
	Avg       float64 `json:"avg"`
	VoteCount int     `json:"vote_count"`
	ProductID string  `json:"product_id"`
}
