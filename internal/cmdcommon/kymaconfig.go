package cmdcommon

import "context"

// KymaConfig contains data common for all subcommands
type KymaConfig struct {
	Ctx context.Context
}
