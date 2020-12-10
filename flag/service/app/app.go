package app

type App struct {
	// ExtraAnnotations allows adding extra annotations to each created AppCR.
	// Annotations are supplied as a string slice using "key:value,key2:value2"
	// format.
	ExtraAnnotations string
}
