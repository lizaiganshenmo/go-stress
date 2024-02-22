package node

import "context"

type WorkNodeDriver interface {
	BaseNodeDriver

	Work(ctx context.Context)
}
