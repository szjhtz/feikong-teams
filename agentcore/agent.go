package agentcore

type Agent interface {
	Name() string
	Description() string
	RuntimeAgent() any
}

type runtimeAgent struct {
	name        string
	description string
	runtime     any
}

func WrapRuntimeAgent(name, description string, runtime any) Agent {
	return &runtimeAgent{name: name, description: description, runtime: runtime}
}

func (a *runtimeAgent) Name() string {
	if a == nil {
		return ""
	}
	return a.name
}

func (a *runtimeAgent) Description() string {
	if a == nil {
		return ""
	}
	return a.description
}

func (a *runtimeAgent) RuntimeAgent() any {
	if a == nil {
		return nil
	}
	return a.runtime
}
