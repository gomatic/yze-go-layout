package layout_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"

	layout "github.com/gomatic/yze-go-layout"
)

func TestLayoutCorrespondence(t *testing.T) {
	analysistest.Run(
		t, analysistest.TestData(), layout.Analyzer,
		"m/internal/app/commands/greet",
		"m/internal/app/commands/greet/flags",
		"m/internal/domain/greet",
		"m/internal/domain/greet/deep",
		"m/internal/app/commands/orphan",
		"m/internal/app/commands/stub",
		"m/internal/app/commands/hollow",
		"m/internal/domain/lonely",
		"m/internal/util",
		"m/examples",
	)
}

func TestRegistrationIsWellFormed(t *testing.T) {
	assert.NoError(t, layout.Registration.Validate())
	assert.Equal(t, "yze/layout", layout.Registration.RuleID())
	assert.Same(t, layout.Analyzer, layout.Registration.Analyzer)
}
