package greeting_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/vearutop/cache-story/internal/domain/greeting"
	"testing"
)

func TestSimpleMaker_GreetingMaker(t *testing.T) {
	sm := &greeting.SimpleMaker{}

	assert.Equal(t, sm, sm.GreetingMaker())
}
