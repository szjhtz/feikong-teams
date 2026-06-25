package runtime

import (
	"context"
	"fkteams/internal/domain/message"
)

type SteeringSource func(context.Context) ([]message.Message, error)

type steeringSourceKey struct{}

func WithSteeringSource(ctx context.Context, source SteeringSource) context.Context {
	if source == nil {
		return ctx
	}
	return context.WithValue(ctx, steeringSourceKey{}, source)
}

func SteeringSourceFromContext(ctx context.Context) (SteeringSource, bool) {
	source, ok := ctx.Value(steeringSourceKey{}).(SteeringSource)
	return source, ok
}
