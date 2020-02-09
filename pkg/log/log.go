package log

import (
	"context"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

//Set up with  Controller based on the level

//Get the request id

//New function sets the logging level based on the flag and also sets it with controller
func New(debug ...bool) {
	enabled := false
	if len(debug) == 0 {
		enabled = true
	} else {
		enabled = debug[0]
	}
	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = enabled
	}))
}

//Logger with
func Logger(ctx context.Context, names ...string) logr.Logger {
	logk := ctrl.Log.Logger
	for _, name := range names {
		logk = logk.WithName(name)
	}
	rId := ctx.Value("request_id")
	if rId != nil {
		logk = logk.WithValues("request_id", rId)
	}

	return logk
}
