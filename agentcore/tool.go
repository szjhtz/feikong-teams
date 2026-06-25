package agentcore

import runtimeport "fkteams/internal/ports/runtime"

type ToolInfo = runtimeport.ToolInfo
type ToolInvocation = runtimeport.ToolInvocation
type ToolRuntimeMetadata = runtimeport.ToolRuntimeMetadata
type ToolResult = runtimeport.ToolResult
type Tool = runtimeport.Tool
type ToolInputTypeProvider = runtimeport.ToolInputTypeProvider

var WithToolRuntimeMetadata = runtimeport.WithToolRuntimeMetadata
var ToolRuntimeMetadataFromContext = runtimeport.ToolRuntimeMetadataFromContext
var InferTool = runtimeport.InferTool
var NewTool = runtimeport.NewTool
