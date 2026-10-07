// Package logger is a thin wrapper over log/slog that emits structured JSON logs
// enriched with correlation fields pulled from context (trace_id, user_id) and the
// call site (source). Initialize it once at startup with Init; it is safe to use
// before Init (it defaults to info-level JSON on stdout).
package logger

import (
	"context"
	"log/slog"
	"os"

	"McQueens_Tea_Cup/pkg/tracer"
)

const (
	EnvDevelopment = "dev"
	EnvProduction  = "prod"
)

type SLogger struct {
	logger slog.Logger
}

// appLogger has a safe default so logging before Init does not panic.
var appLogger = newSLogger(EnvProduction, slog.LevelInfo)

func newSLogger(env string, level slog.Level) SLogger {
	var handler slog.Handler
	if env == EnvDevelopment {
		// Dev: colored, human-readable console output.
		handler = newPrettyHandler(os.Stdout, level)
	} else {
		// Prod: structured JSON for log ingestion.
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     level,
			AddSource: false, // source is captured manually via getCaller
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if len(groups) == 0 && a.Key == slog.TimeKey {
					a.Key = "timestamp"
				}
				return a
			},
		})
	}
	return SLogger{logger: *slog.New(handler)}
}

// Init configures the global logger from the environment: "dev" => colored,
// debug-level console output; anything else => info-level JSON.
func Init(env string) {
	level := slog.LevelInfo
	if env == EnvDevelopment {
		level = slog.LevelDebug
	}
	appLogger = newSLogger(env, level)
}

// contextAttrs returns the standard correlation attributes pulled from context.
// Missing values are emitted as empty strings so the log schema stays consistent.
func contextAttrs(ctx context.Context) []any {
	return []any{
		slog.String("trace_id", tracer.TraceIDFromContext(ctx)),
		slog.String("user_id", tracer.UserIDFromContext(ctx)),
	}
}

func Debug(ctx context.Context, msg string, args ...any) {
	allAttrs := append(contextAttrs(ctx), slog.String("source", getCaller(3)))
	allAttrs = append(allAttrs, args...)
	appLogger.logger.DebugContext(ctx, msg, allAttrs...)
}

func Info(ctx context.Context, msg string, args ...any) {
	allAttrs := append(contextAttrs(ctx), slog.String("source", getCaller(3)))
	allAttrs = append(allAttrs, args...)
	appLogger.logger.InfoContext(ctx, msg, allAttrs...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	allAttrs := append(contextAttrs(ctx), slog.String("source", getCaller(3)))
	allAttrs = append(allAttrs, args...)
	appLogger.logger.WarnContext(ctx, msg, allAttrs...)
}

func Error(ctx context.Context, msg string, err error, args ...any) {
	stackTrace, errorCode := getErrorStack(err)

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	allAttrs := append(contextAttrs(ctx),
		slog.String("source", getCaller(3)),
		slog.Any("error", errMsg),
		slog.String("stack_trace", stackTrace),
	)
	if errorCode != "" {
		allAttrs = append(allAttrs, slog.String("error_code", errorCode))
	}
	allAttrs = append(allAttrs, args...)

	appLogger.logger.ErrorContext(ctx, msg, allAttrs...)
}
