package voting

type VoteRequest struct {
	ProductID string `json:"product_id"`
	Score     int    `json:"score" validate:"min=1,max=5"`
}
