// Command tfdrift is the agentless Terraform drift detector CLI. This entrypoint
// only wires the command tree; all logic lives in internal/cli and below.
package main

import (
	"os"

	"github.com/yashyaadav/tf-drift-detector/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
