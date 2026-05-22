package telegram

// MessageSender defines the interface for sending Telegram messages.
type MessageSender interface {
	SendMessage(chatID int64, text string) error
}

// PipelineMessageSender extends MessageSender with pipeline message update support.
type PipelineMessageSender interface {
	MessageSender
	SendOrUpdatePipelineMessage(chatID int64, pipelineURL string, text string) error
}
