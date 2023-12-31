package audio

import (
	"strings"

	"github.com/gen2brain/malgo"
)

var (
	malgoContext *malgo.AllocatedContext
)

func initContext() {
	contextConfig := malgo.ContextConfig{}
	contextConfig.ThreadPriority = malgo.ThreadPriority(audioContextConfig.GetInt("ThreadPriority"))

	ctx, err := malgo.InitContext(nil, contextConfig, func(message string) {
		logger.Tracef("[Engine] %v", strings.TrimRight(message, "\n"))
	})
	if err != nil {
		logger.Error("Context 초기화 실패 err=%w", err)
	}

	malgoContext = ctx
}

func disposeContext() {
	defer func() {
		_ = malgoContext.Uninit()
		malgoContext.Free()
	}()
}
