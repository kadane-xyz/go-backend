package logger

import (
	"fmt"
	"sync"

	"slices"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"kadane.xyz/go-backend/v2/internal/config"
)

// Sink types
type SinkType int

const (
	SinkTypeConsole SinkType = iota
	SinkTypeAWS
	SinkTypeDatadog
)

// Log Types
type LogType string

const (
	LogTypeConsole  LogType = "console"
	LogTypeAWS      LogType = "aws"
	LogTypeDatabase LogType = "database"
	LogTypeRedis    LogType = "redis"
	LogTypeJudge0   LogType = "judge0"
	LogTypeFirebase LogType = "firebase"
)

// Log levels to match RFC 5424
type LogLevel int

const (
	LogLevelEmergency LogLevel = iota
	LogLevelAlert
	LogLevelCritical
	LogLevelError
	LogLevelWarning
	LogLevelNotice
	LogLevelInformational
	LogLevelDebug // handled by DEBUG env var ?
)

// Log formats
type LogFormat int

const (
	LogFormatConsole LogFormat = iota
	LogFormatJSON
)

type LoggerSink struct {
	Name         string               // Name of logging sink
	Type         SinkType             // Type of logging sink
	Environments []config.Environment // Application environments to run under
	LogTypes     []LogType            // Types of logs to handle
	LogLevels    []LogLevel           // Log levels supported
	LogFormat    LogFormat            // Log format to use
}

// Declaration of sinks to initialize
var LoggerSinks = []LoggerSink{
	{
		Name: "Console",
		Type: SinkTypeConsole,
		Environments: []config.Environment{
			config.EnvProduction,
			config.EnvStaging,
			config.EnvDevelopment,
			config.EnvTest,
		},
		LogTypes: []LogType{
			LogTypeConsole,
		},
		LogLevels: []LogLevel{
			LogLevelAlert,
			LogLevelCritical,
			LogLevelEmergency,
			LogLevelError,
			LogLevelInformational,
			LogLevelNotice,
			LogLevelWarning,
		},
		LogFormat: LogFormatConsole,
	},
	/*{
		Name: "ConsoleJSON",
		Type: SinkTypeConsole,
		Environments: []config.Environment{
			config.EnvProduction,
			config.EnvStaging,
			config.EnvDevelopment,
			config.EnvTest,
		},
		LogTypes: []LogType{
			LogTypeConsole,
		},
		LogLevels: []LogLevel{
			LogLevelAlert,
			LogLevelCritical,
			LogLevelDebug,
			LogLevelEmergency,
			LogLevelError,
			LogLevelInformational,
			LogLevelNotice,
			LogLevelWarning,
		},
		LogFormat: LogFormatJSON,
	},*/
}

// LoggerSinkClient is a struct that contains the clients for each logger sink
type LoggerSinkClient struct {
	ConsoleLogger map[string]*zap.Logger
	GoogleLogger  map[string]*zap.Logger
	SentryLogger  map[string]*zap.Logger
}

type LoggerSinkManager struct {
	Clients     LoggerSinkClient
	ActiveSinks []LoggerSink
	Config      *config.Config
}

type LoggerMessage struct {
	Level   LogLevel    `json:"level"`
	Message string      `json:"message"`
	Fields  []zap.Field `json:"fields"`
}

func NewLoggerSinkManager(cfg *config.Config) (*LoggerSinkManager, error) {
	manager := &LoggerSinkManager{
		Clients: LoggerSinkClient{
			ConsoleLogger: make(map[string]*zap.Logger),
		},
		ActiveSinks: LoggerSinks,
		Config:      cfg,
	}

	if err := manager.InitializeLoggerSinks(); err != nil {
		return nil, err
	}

	return manager, nil
}

// Validate the logger sinks
func validateLoggerSinks(sinks []LoggerSink) error {
	for _, sink := range sinks {
		if sink.Name == "" {
			return fmt.Errorf("logger sink name is empty")
		}

		if len(sink.Environments) == 0 {
			return fmt.Errorf("logger sink environments are empty")
		}

		if len(sink.LogTypes) == 0 {
			return fmt.Errorf("logger sink log types are empty")
		}

		if len(sink.LogLevels) == 0 {
			return fmt.Errorf("logger sink log levels are empty")
		}

		for _, logType := range sink.LogTypes {
			if !containsLogType(sink.LogTypes, logType) {
				return fmt.Errorf("logger sink log type is invalid: %v", logType)
			}
		}

		for _, logLevel := range sink.LogLevels {
			if !containsLogLevel(sink.LogLevels, logLevel) {
				return fmt.Errorf("logger sink log level is invalid: %v", logLevel)
			}
		}
	}

	return nil
}

// NewLoggerSinkManagerForTest creates a new logger sink manager for testing
func NewLoggerSinkManagerForTest() *LoggerSinkManager {
	manager := &LoggerSinkManager{
		ActiveSinks: []LoggerSink{},
		Config:      &config.Config{},
	}

	manager.InitializeTestLoggerSinks()

	return manager
}

// Initialize logger sinks for testing
func (m *LoggerSinkManager) InitializeTestLoggerSinks() error {
	m.ActiveSinks = []LoggerSink{
		{
			Name: "Test",
			Type: SinkTypeConsole,
		},
	}

	return nil
}

// Initialize logger sinks for application logging
func (m *LoggerSinkManager) InitializeLoggerSinks() error {
	// Validate the logger sinks
	if err := validateLoggerSinks(m.ActiveSinks); err != nil {
		return err
	}

	// If debug is enabled, add a debug sink
	if m.Config.Debug {
		m.ActiveSinks = append(m.ActiveSinks, LoggerSink{
			Name: "Debug",
			Type: SinkTypeConsole,
			Environments: []config.Environment{
				config.EnvDevelopment,
				config.EnvTest,
			},
			LogTypes: []LogType{
				LogTypeConsole,
			},
			LogLevels: []LogLevel{
				LogLevelDebug,
			},
			LogFormat: LogFormatConsole,
		})
	}

	for _, sink := range m.ActiveSinks {
		switch sink.Type {
		case SinkTypeConsole:
			if err := m.initializeConsoleSink(sink); err != nil {
				return nil
			}
		}
	}

	return nil
}

// Map Log levels to zap log levels
func getZapConfig(environment config.Environment) *zap.Config {
	var zapConfig zap.Config

	switch environment {
	case config.EnvProduction, config.EnvStaging:
		zapConfig = zap.NewProductionConfig()
	case config.EnvDevelopment, config.EnvTest:
		zapConfig = zap.NewDevelopmentConfig()
	default:
		return nil
	}

	return &zapConfig
}

// Map Log levels to zap log levels
func getZapLogLevel(levels []LogLevel) zap.AtomicLevel {
	// If no levels specified, default to Info
	if len(levels) == 0 {
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Find the most verbose level (highest numeric value in our enum)
	var mostVerbose LogLevel = LogLevelEmergency // Start with least verbose
	for _, level := range levels {
		if level > mostVerbose {
			mostVerbose = level
		}
	}

	var zapLevel zapcore.Level
	switch mostVerbose {
	case LogLevelDebug:
		zapLevel = zapcore.DebugLevel
	case LogLevelInformational:
		zapLevel = zapcore.InfoLevel
	case LogLevelNotice:
		zapLevel = zapcore.InfoLevel // Zap has no direct equivalent
	case LogLevelWarning:
		zapLevel = zapcore.WarnLevel
	case LogLevelError:
		zapLevel = zapcore.ErrorLevel
	case LogLevelCritical, LogLevelAlert, LogLevelEmergency:
		zapLevel = zapcore.DPanicLevel // Development panic - logs and then panics (only in development)
	default:
		zapLevel = zapcore.InfoLevel
	}

	return zap.NewAtomicLevelAt(zapLevel)
}

func getZapEncodingOutputPaths(sink LoggerSink) []string {
	switch sink.Type {
	case SinkTypeConsole:
		return []string{"stdout"}
	case SinkTypeGoogle:
		return []string{"stdout"}
	case SinkTypeSentry:
		return []string{"stdout"}
	default:
		return nil
	}
}

func getZapEncoding(format LogFormat) (*zap.Config, error) {
	var config zap.Config

	switch format {
	case LogFormatConsole:
		config.Encoding = "console"
	case LogFormatJSON:
		config.Encoding = "json"
	default:
		return nil, fmt.Errorf("invalid sink log format: %v", format)
	}

	return &config, nil
}

func buildZapLogger(config *zap.Config) (*zap.Logger, error) {
	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func initializeZapLogger(environment config.Environment, sink LoggerSink) (*zap.Logger, error) {
	config := getZapConfig(environment)

	// Configure zap logger
	config.Level = getZapLogLevel(sink.LogLevels)

	// Get zap configuration logger using appropriate application environment
	encoding, err := getZapEncoding(sink.LogFormat)
	if err != nil {
		return nil, err
	}

	// Configure zap logger
	config.Encoding = encoding.Encoding

	// Set encoder configs for all formats
	config.EncoderConfig.LevelKey = "level"     // Was "L"
	config.EncoderConfig.TimeKey = "time"       // Already correct
	config.EncoderConfig.CallerKey = "file"     // Was "C"
	config.EncoderConfig.MessageKey = "message" // Was "M"
	config.EncoderConfig.StacktraceKey = "stacktrace"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// Only set color for console logs
	if sink.LogFormat == LogFormatConsole {
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	if sink.LogFormat == LogFormatJSON {
		config.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	}

	// Configure zap logger
	config.OutputPaths = getZapEncodingOutputPaths(sink)
	config.ErrorOutputPaths = config.OutputPaths

	// Build zap logger
	return buildZapLogger(config)
}

func (m *LoggerSinkManager) initializeConsoleSink(sink LoggerSink) error {
	// Initialize zap logger
	logger, err := initializeZapLogger(m.Config.Environment, sink)
	if err != nil {
		return err
	}

	// Set console logger instance
	m.Clients.ConsoleLogger[sink.Name] = logger

	return nil
}

func (m *LoggerSinkManager) initializeGoogleSink(sink LoggerSink) error {
	// Initialize zap logger
	logger, err := initializeZapLogger(m.Config.Environment, sink)
	if err != nil {
		return err
	}

	// Set google logger instance
	m.Clients.GoogleLogger[sink.Name] = logger

	return nil
}

func (m *LoggerSinkManager) initializeSentrySink(sink LoggerSink) error {
	// Initialize zap logger
	logger, err := initializeZapLogger(m.Config.Environment, sink)
	if err != nil {
		return err
	}

	// Set sentry logger instance
	m.Clients.SentryLogger[sink.Name] = logger

	return nil
}

// containsLogLevel checks if a target log level exists in a slice of log levels
func containsLogLevel(levels []LogLevel, target LogLevel) bool {
	return slices.Contains(levels, target)
}

// containsLogType checks if a target log type exists in a slice of log types
func containsLogType(types []LogType, target LogType) bool {
	return slices.Contains(types, target)
}

// LoggerSinkManager methods for different log levels
func (m *LoggerSinkManager) Debug(logType LogType, msg string, fields ...zap.Field) {
	m.processAndSendLog(LogLevelDebug, logType, msg, fields...)
}

func (m *LoggerSinkManager) Info(logType LogType, msg string, fields ...zap.Field) {
	m.processAndSendLog(LogLevelInformational, logType, msg, fields...)
}

func (m *LoggerSinkManager) Notice(logType LogType, msg string, fields ...zap.Field) {
	m.processAndSendLog(LogLevelNotice, logType, msg, fields...)
}

func (m *LoggerSinkManager) Warn(logType LogType, msg string, fields ...zap.Field) {
	m.processAndSendLog(LogLevelWarning, logType, msg, fields...)
}

func (m *LoggerSinkManager) Error(logType LogType, msg string, fields ...zap.Field) {
	m.processAndSendLog(LogLevelError, logType, msg, fields...)
}

func (m *LoggerSinkManager) Critical(logType LogType, msg string, fields ...zap.Field) {
	m.processAndSendLog(LogLevelCritical, logType, msg, fields...)
}

func (m *LoggerSinkManager) Alert(logType LogType, msg string, fields ...zap.Field) {
	m.processAndSendLog(LogLevelAlert, logType, msg, fields...)
}

func (m *LoggerSinkManager) Emergency(logType LogType, msg string, fields ...zap.Field) {
	m.processAndSendLog(LogLevelEmergency, logType, msg, fields...)
}

// Shorthand convenience methods for common log types
func (m *LoggerSinkManager) ConsoleDebug(msg string, fields ...zap.Field) {
	m.Debug(LogTypeConsole, msg, fields...)
}

func (m *LoggerSinkManager) RequestInfo(msg string, fields ...zap.Field) {
	m.Info(LogTypeRequest, msg, fields...)
}

func (m *LoggerSinkManager) JobError(msg string, fields ...zap.Field) {
	m.Error(LogTypeJobs, msg, fields...)
}

func formatLogMessage(logType LogType, fields ...zap.Field) []zap.Field {
	// Add extra data to the fields
	var extraFields []zap.Field
	extraFields = append(extraFields, zap.String("type", string(logType)))

	return append(fields, extraFields...)
}

// Core processing and distribution logic
func (m *LoggerSinkManager) processAndSendLog(level LogLevel, logType LogType, message string, fields ...zap.Field) {
	// Filter sinks that handle this log type and level
	var applicableSinks []LoggerSink
	for _, sink := range m.ActiveSinks {
		if containsLogType(sink.LogTypes, logType) && containsLogLevel(sink.LogLevels, level) {
			applicableSinks = append(applicableSinks, sink)
		}
	}

	// If no sinks handle this message return early
	if len(applicableSinks) == 0 {
		return
	}

	// Format the log message
	fields = formatLogMessage(logType, fields...)

	// Create wait group for concurrent sink writes
	var wg sync.WaitGroup
	wg.Add(len(applicableSinks))

	// Send to each sink concurrently
	for _, sink := range applicableSinks {
		go func(s LoggerSink) {
			defer wg.Done()
			m.writeLogToSink(s, level, message, fields...)
		}(sink)
	}

	// Wait for all logs to be written
	wg.Wait()
}

// Sends log to the appropriate sink
func (m *LoggerSinkManager) writeLogToSink(sink LoggerSink, level LogLevel, message string, fields ...zap.Field) {
	switch sink.Type {
	case SinkTypeConsole:
		if m.Clients.ConsoleLogger != nil {
			writeLogWithLevel(m.Clients.ConsoleLogger[sink.Name], level, message, fields...)
		}
	case SinkTypeGoogle:
		if m.Clients.GoogleLogger != nil {
			writeLogWithLevel(m.Clients.GoogleLogger[sink.Name], level, message, fields...)
		}
	case SinkTypeSentry:
		if m.Clients.SentryLogger != nil {
			writeLogWithLevel(m.Clients.SentryLogger[sink.Name], level, message, fields...)
		}
	}
}

// Maps our log level to zap logging methods
func writeLogWithLevel(logger *zap.Logger, level LogLevel, msg string, fields ...zap.Field) {
	switch level {
	case LogLevelDebug:
		logger.Debug(msg, fields...)
	case LogLevelInformational, LogLevelNotice:
		logger.Info(msg, fields...)
	case LogLevelWarning:
		logger.Warn(msg, fields...)
	case LogLevelError:
		logger.Error(msg, fields...)
	case LogLevelCritical:
		logger.DPanic(msg, fields...)
	case LogLevelAlert, LogLevelEmergency:
		logger.Panic(msg, fields...) // Highest severity in zap
	}
}

func (m *LoggerSinkManager) Log(level LogLevel, logType LogType, message string, fields ...zap.Field) {
	m.processAndSendLog(level, logType, message, fields...)
}
