package agentcore

import runtimeport "fkteams/internal/ports/runtime"

type Agent = runtimeport.Agent
type AgentToolNameFunc = runtimeport.AgentToolNameFunc
type AgentToolDisplayFunc = runtimeport.AgentToolDisplayFunc
type AgentToolConfig = runtimeport.AgentToolConfig
type UnknownToolHandler = runtimeport.UnknownToolHandler
type RetryContext = runtimeport.RetryContext
type RetryDecision = runtimeport.RetryDecision
type ModelRetryConfig = runtimeport.ModelRetryConfig
type ChatAgentConfig = runtimeport.ChatAgentConfig
type LoopAgentConfig = runtimeport.LoopAgentConfig
type DeepAgentConfig = runtimeport.DeepAgentConfig
type RunnerConfig = runtimeport.RunnerConfig
type SummaryPersistCallback = runtimeport.SummaryPersistCallback

const DefaultMaxTokensBeforeSummary = runtimeport.DefaultMaxTokensBeforeSummary

var WithSummaryPersistCallback = runtimeport.WithSummaryPersistCallback
var SummaryPersistCallbackFromContext = runtimeport.SummaryPersistCallbackFromContext

type SummaryConfig = runtimeport.SummaryConfig
type DispatchConfig = runtimeport.DispatchConfig
type Engine = runtimeport.Engine
type RuntimeInfo = runtimeport.RuntimeInfo
type RuntimeHealth = runtimeport.RuntimeHealth
type RuntimeInspector = runtimeport.RuntimeInspector
type ModelDecorator = runtimeport.ModelDecorator
type AgentPipelineProvider = runtimeport.AgentPipelineProvider
type ToolPipelineProvider = runtimeport.ToolPipelineProvider
type MCPToolProvider = runtimeport.MCPToolProvider
