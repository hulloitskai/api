package gqlutil

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cockroachdb/errors"
	"github.com/vektah/gqlparser/gqlerror"

	"go.stevenxie.me/gopkg/zero"
)

// PresentError is a graphql.ErrorPresenterFunc that knows how to present
// errors created by cockroachdb/errors.
func PresentError(ctx context.Context, err error) *gqlerror.Error {
	exts := make(map[string]zero.Interface)

	// Append to extensions.
	if hints := errors.GetAllHints(err); len(hints) > 0 {
		exts["hints"] = hints
	}
	if links := errors.GetAllIssueLinks(err); len(links) > 0 {
		exts["issueLinks"] = links
	}
	if cause := errors.UnwrapAll(err); (cause != err) && (cause != nil) {
		if msg := cause.Error(); msg != err.Error() {
			exts["cause"] = msg
		}
	}
	if trace := errors.GetReportableStackTrace(err); trace != nil {
		exts["culprit"] = trace.Culprit()
	}
	if file, line, fn, ok := getOneLineSource(err); ok {
		exts["source"] = struct {
			File string `json:"file"`
			Line int    `json:"line"`
			Fn   string `json:"fn"`
		}{file, line, fn}
	}

	gqlErr := graphql.DefaultErrorPresenter(ctx, err)
	if len(exts) > 0 {
		if len(gqlErr.Extensions) == 0 {
			gqlErr.Extensions = exts
		} else {
			for k, v := range exts {
				gqlErr.Extensions[k] = v
			}
		}
	}
	return gqlErr
}

var _ graphql.ErrorPresenterFunc = PresentError

// // ResultErrorCallback creates a handler.ResultCallbackFn that logs errors with
// // log.
// func ResultErrorCallback(log *logrus.Entry) handler.ResultCallbackFn {
// 	return func(
// 		ctx context.Context,
// 		ps *graphql.Params, res *graphql.Result,
// 		_ []byte,
// 	) {
// 		if res.HasErrors() {
// 			for _, err := range res.Errors {
// 				entry := log.
// 					WithError(err).
// 					WithContext(ctx).
// 					WithFields(logrus.Fields{
// 						"request":   ps.RequestString,
// 						"variables": ps.VariableValues,
// 					})

// 				{
// 					var gqlErr *gqlerrors.Error
// 					if errors.As(err.OriginalError(), &gqlErr) {
// 						if gqlErr.OriginalError == nil {
// 							// No original resolver error; assume that this error results
// 							// from an invalid request.
// 							entry.Debug("Error while parsing query.")
// 						} else {
// 							entry.Error("Error while resolving query.")
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// }

// // FormatError formats an GraphQL error, using additional context gleaned from
// // package cockroachdb/errors.
// func FormatError(err error) gqlerrors.FormattedError {
// 	fmtErr := gqlerrors.FormatError(err)

// 	// Read original error.
// 	{
// 		var gqlErr *gqlerrors.Error
// 		if errors.As(err, &gqlErr) {
// 			err = gqlErr.OriginalError
// 		}
// 	}

// 	// Init extensions.
// 	if fmtErr.Extensions == nil {
// 		fmtErr.Extensions = make(map[string]zero.Interface)
// 	}

// 	// Append to extensions.
// 	if hints := errors.GetAllHints(err); len(hints) > 0 {
// 		fmtErr.Extensions["hints"] = hints
// 	}
// 	if links := errors.GetAllIssueLinks(err); len(links) > 0 {
// 		fmtErr.Extensions["issueLinks"] = links
// 	}
// 	if cause := errors.UnwrapAll(err); (cause != err) && (cause != nil) {
// 		if msg := cause.Error(); msg != err.Error() {
// 			fmtErr.Extensions["cause"] = msg
// 		}
// 	}
// 	if trace := errors.GetReportableStackTrace(err); trace != nil {
// 		fmtErr.Extensions["culprit"] = trace.Culprit()
// 	}
// 	if file, line, fn, ok := getOneLineSource(err); ok {
// 		fmtErr.Extensions["source"] = struct {
// 			File string `json:"file"`
// 			Line int    `json:"line"`
// 			Fn   string `json:"fn"`
// 		}{file, line, fn}
// 	}

// 	return fmtErr
// }
