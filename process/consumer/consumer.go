package consumer

// The definiation of API consumer information
type ConsumerInfo struct {
	UserID string
}

// The consumer information according to token
func GetTokenInfo(token string) (*ConsumerInfo, error) {
	// To simply design, just using token as UserID here
	consumerInfo := ConsumerInfo{
		UserID: token,
	}

	return &consumerInfo, nil
}
