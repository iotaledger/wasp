package smGPA

type (
	brcIsValidFun          func() bool
	brcRequestCompletedFun func()
)

type blockRequestCallbackImpl struct {
	isValidFun          brcIsValidFun
	requestCompletedFun brcRequestCompletedFun
}

var (
	brcAlwaysValidFun            = func() bool { return true }
	brcIgnoreRequestCompletedFun = func() {}
)

var (
	_ blockRequestCallback   = &blockRequestCallbackImpl{}
	_ brcIsValidFun          = brcAlwaysValidFun
	_ brcRequestCompletedFun = brcIgnoreRequestCompletedFun
)

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
