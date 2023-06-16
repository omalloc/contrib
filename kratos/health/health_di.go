package health

import (
	"go.uber.org/fx"
)

// AsChecker is a health-checker transformer.
func AsChecker[T Checker]() interface{} {
	return fx.Annotate(
		func(typo T) Checker {
			return typo
		},
		fx.ResultTags(`group:"health"`),
	)
}

// AsHealth is a health-server transformer.
func AsHealth(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(`group:"health"`),
	)
}
