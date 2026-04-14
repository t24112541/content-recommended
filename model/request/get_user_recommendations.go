package request

type GetUserRecommendations struct {
	UserId int64 `params:"user_id" validate:"required,gt=0"`
	Limit  int   `query:"limit"`
}
