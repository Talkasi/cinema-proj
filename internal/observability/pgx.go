package observability

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/tracelog"
)

type pgxTraceLogger struct{}

func (pgxTraceLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	if !activeConfig.TracesEnabled {
		return
	}

	logLine := msg
	if len(data) > 0 {
		logLine = logLine + " data=" + fmt.Sprint(data)
	}

	log.Printf("pgx %s: %s", level, logLine)
}

// ConfigurePgxTracing instruments pgx connection only when tracing is enabled.
func ConfigurePgxTracing(connConfig *pgx.ConnConfig) {
	if connConfig == nil {
		return
	}

	if !activeConfig.TracesEnabled {
		return
	}

	connConfig.Tracer = &tracelog.TraceLog{
		Logger:   pgxTraceLogger{},
		LogLevel: tracelog.LogLevelDebug,
	}
}
