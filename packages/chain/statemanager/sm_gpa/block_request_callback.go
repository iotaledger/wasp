package sm_gpa

type (
	brcIsValidFun          func() bool
	brcRequestCompletedFun func()
)

type blockRequestCallbackImpl struct {
	isValidFun          brcIsValidFun
	requestCompletedFun brcRequestCompletedFun
}

var _ blockRequestCallback = &blockRequestCallbackImpl{}

func newBlockRequestCallback(ivf brcIsValidFun, rcf brcRequestCompletedFun) blockRequestCallback {
	return &blockRequestCallbackImpl{
		isValidFun:          ivf,
		requestCompletedFun: rcf,
	}
}

func (brciT *blockRequestCallbackImpl) isValid() bool {
	return brciT.isValidFun()
}

func (brciT *blockRequestCallbackImpl) requestCompleted() {
	brciT.requestCompletedFun()
}
