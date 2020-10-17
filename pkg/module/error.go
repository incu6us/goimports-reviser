package module

type UndefinedModuleError struct{}

func (e *UndefinedModuleError) Error() string {
	return "module is undefined"
}

type PathIsNotSetError struct{}

func (e *PathIsNotSetError) Error() string {
	return "path is not set"
}
