package agentcore

type ChatModel interface {
	RuntimeModel() any
}

type runtimeChatModel struct {
	runtime any
}

func WrapRuntimeChatModel(runtime any) ChatModel {
	return &runtimeChatModel{runtime: runtime}
}

func (m *runtimeChatModel) RuntimeModel() any {
	if m == nil {
		return nil
	}
	return m.runtime
}
