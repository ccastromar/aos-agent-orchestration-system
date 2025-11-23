package runtime

import (
	"github.com/ccastromar/aos-banking-v2/internal/llm"
)

type Runtime struct {
	SpecsLoaded bool
	LLMClient   llm.LLMClient
}
