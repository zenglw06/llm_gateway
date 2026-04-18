package logger

import (
    "os"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var (
    logger *zap.Logger
    sugar  *zap.SugaredLogger
)

// Config 日志配置
type Config struct {
    Level      string `mapstructure:"level"`
    Format     string `mapstructure:"format"`
    OutputPath string `mapstructure:"output_path"`
    Debug      bool   `mapstructure:"debug"`
}

// Init 初始化日志
func Init(cfg *Config) error {
    var level zapcore.Level
    if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
        level = zapcore.InfoLevel
    }

    encoderConfig := zap.NewProductionEncoderConfig()
    encoderConfig.TimeKey = "time"
    encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

    var encoder zapcore.Encoder
    if cfg.Format == "json" {
        encoder = zapcore.NewJSONEncoder(encoderConfig)
    } else {
        encoder = zapcore.NewConsoleEncoder(encoderConfig)
    }

    var output zapcore.WriteSyncer
    if cfg.OutputPath == "stdout" || cfg.OutputPath == "" {
        output = zapcore.AddSync(os.Stdout)
    } else {
        file, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
        if err != nil {
            return err
        }
        output = zapcore.AddSync(file)
    }

    core := zapcore.NewCore(encoder, output, level)

    var options []zap.Option
    if cfg.Debug {
        options = append(options, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
    }

    logger = zap.New(core, options...)
    sugar = logger.Sugar()

    return nil
}

// GetLogger 获取zap Logger实例
func GetLogger() *zap.Logger {
    if logger == nil {
        // 默认初始化
        logger, _ = zap.NewProduction()
        sugar = logger.Sugar()
    }
    return logger
}

// GetSugar 获取SugaredLogger实例
func GetSugar() *zap.SugaredLogger {
    if sugar == nil {
        // 默认初始化
        logger, _ = zap.NewProduction()
        sugar = logger.Sugar()
    }
    return sugar
}

// 以下是快捷方法
func Debug(args ...interface{}) {
    GetSugar().Debug(args...)
}

func Debugf(template string, args ...interface{}) {
    GetSugar().Debugf(template, args...)
}

func Info(args ...interface{}) {
    GetSugar().Info(args...)
}

func Infof(template string, args ...interface{}) {
    GetSugar().Infof(template, args...)
}

func Warn(args ...interface{}) {
    GetSugar().Warn(args...)
}

func Warnf(template string, args ...interface{}) {
    GetSugar().Warnf(template, args...)
}

func Error(args ...interface{}) {
    GetSugar().Error(args...)
}

func Errorf(template string, args ...interface{}) {
    GetSugar().Errorf(template, args...)
}

func Fatal(args ...interface{}) {
    GetSugar().Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
    GetSugar().Fatalf(template, args...)
}

func Sync() error {
    if logger != nil {
        return logger.Sync()
    }
    return nil
}
