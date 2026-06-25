package agentcore

import (
	"context"
	runtimeport "fkteams/internal/ports/runtime"
)

type InterruptRuntime = runtimeport.InterruptRuntime
type InterruptMetadata = runtimeport.InterruptMetadata
type InterruptPayload = runtimeport.InterruptPayload

var RegisterInterruptRuntime = runtimeport.RegisterInterruptRuntime
var WithInterruptMetadata = runtimeport.WithInterruptMetadata
var InterruptMetadataFromContext = runtimeport.InterruptMetadataFromContext
var RequestInterrupt = runtimeport.RequestInterrupt
var GetInterruptState = runtimeport.GetInterruptState

func GetResumeContext[T any](ctx context.Context) (bool, bool, T) {
	return runtimeport.GetResumeContext[T](ctx)
}
