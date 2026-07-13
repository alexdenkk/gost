package domain

type AgentRequest struct {
	Message         string      `json:"message"`
	ParentMessageID string      `json:"parent_message_id"`
	FileIDs         []string    `json:"file_ids"`
	Metadata        interface{} `json:"metadata"`
}

type AgentResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
}
