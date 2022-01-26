package errors

var FailedToLoadError ErrorDefinition = NewBlockErrorDefinition(1, "Failed to load %v. Tried to %v with %v")
var CouldNotReadFromError ErrorDefinition = NewBlockErrorDefinition(2, "Could not read from %v")

var GeneralErrorCollection = DefaultErrorCollection{
	Errors: map[int]ErrorDefinition{
		1: FailedToLoadError,
		2: CouldNotReadFromError,
	},
}
