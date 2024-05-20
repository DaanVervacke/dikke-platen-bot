package types

type TelegramBotUpdate struct {
	Message struct {
		ID   int `json:"message_id"`
		Chat struct {
			ID int `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"channel_post"`
}
