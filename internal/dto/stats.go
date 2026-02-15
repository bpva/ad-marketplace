package dto

type ChannelFullInfo struct {
	ParticipantsCount int    `json:"participants_count"`
	LinkedChatID      int64  `json:"linked_chat_id,omitempty"`
	AdminsCount       int    `json:"admins_count"`
	OnlineCount       int    `json:"online_count"`
	About             string `json:"about"`
	CanViewStats      bool   `json:"can_view_stats"`
	StatsDC           int    `json:"stats_dc"`
}
