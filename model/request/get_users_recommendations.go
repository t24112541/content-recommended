package request

type GetUsersRecommendations struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}
