package constants

import "{{.ProjectName}}/app"

// Constants for the app
// import for '.' ， to use the app package
var (
	Taurus  = app.T
	Cleanup = app.Cleanup
	Err     = app.Err
)
