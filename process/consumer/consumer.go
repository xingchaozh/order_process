package consumer

type ConsumerInfo struct {
	UserID string
}

func GetTokenInfo(token string) (*ConsumerInfo, error) {
	// To simply design, just using token as UserID here
	consumerInfo := ConsumerInfo{
		UserID: token,
	}

	return &consumerInfo, nil
}
