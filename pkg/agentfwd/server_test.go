package agentfwd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProbeSSHAgent(t *testing.T) {
	err := TestSSHAgentConnectivity()
	assert.NoError(t, err)
}
