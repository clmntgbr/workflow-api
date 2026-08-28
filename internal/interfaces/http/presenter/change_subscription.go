package presenter

import cmdsubscription "go-api/internal/application/command/subscription"

type ChangeSubscriptionResponse struct {
	URL     *string `json:"url"`
	Updated bool    `json:"updated"`
}

func NewChangeSubscriptionResponse(result *cmdsubscription.CreateSubscriptionResult) ChangeSubscriptionResponse {
	if result == nil {
		return ChangeSubscriptionResponse{}
	}
	return ChangeSubscriptionResponse{
		URL:     optionalNonEmptyString(result.URL),
		Updated: result.Updated,
	}
}

type CheckoutSessionResponse struct {
	URL string `json:"url"`
}

func NewCheckoutSessionResponse(url string) CheckoutSessionResponse {
	return CheckoutSessionResponse{URL: url}
}
