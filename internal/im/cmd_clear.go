package im

import "context"

// ClearCommand implements /clear.
// It soft-deletes the current ChannelSession and clears the LLM context so
// the next message starts a completely fresh conversation.
type ClearCommand struct{}

func newClearCommand() *ClearCommand { return &ClearCommand{} }

func (c *ClearCommand) Name() string { return "clear" }
func (c *ClearCommand) Description() string {
	return "Xóa trí nhớ cuộc trò chuyện, tin nhắn tiếp theo sẽ bắt đầu phiên hoàn toàn mới"
}

func (c *ClearCommand) Execute(_ context.Context, _ *CommandContext, _ []string) (*CommandResult, error) {
	return &CommandResult{
		Content: "✅ Đã xóa cuộc trò chuyện, tin nhắn tiếp theo sẽ bắt đầu phiên hoàn toàn mới.",
		Action:  ActionClear,
	}, nil
}
