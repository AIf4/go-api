package logger

import (
	"os"
	"path/filepath"
	"time"

	"go-meli/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(cfg *config.Config) (*zap.Logger, error) {
	// --- encoder: cómo se formatean los logs ---
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
	}

	// --- nivel mínimo según entorno ---
	level := zapcore.InfoLevel
	if !cfg.IsProd() {
		level = zapcore.DebugLevel
	}

	// --- destinos de escritura ---
	cores := []zapcore.Core{
		consoleCore(encoderCfg, level, cfg.IsProd()),
		fileCore(encoderCfg, cfg.LogDir),
		errorFileCore(encoderCfg, cfg.LogDir),
	}

	log := zap.New(
		zapcore.NewTee(cores...), // escribe en todos los destinos simultáneamente
		zap.AddCaller(),          // agrega archivo y línea al log
	)

	return log, nil
}

// consoleCore — escribe en stdout
// producción: JSON puro / desarrollo: formato legible con colores
func consoleCore(enc zapcore.EncoderConfig, level zapcore.Level, isProd bool) zapcore.Core {
	var encoder zapcore.Encoder
	if isProd {
		encoder = zapcore.NewJSONEncoder(enc)
	} else {
		enc.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(enc)
	}

	return zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)
}

// fileCore — archivo diario con rotación, todos los niveles
func fileCore(enc zapcore.EncoderConfig, logDir string) zapcore.Core {
	fileName := filepath.Join(logDir, "app-"+dailyStamp()+".log")

	rotator := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    100,  // MB antes de rotar
		MaxBackups: 30,   // máximo 30 archivos de backup
		MaxAge:     30,   // días que se conservan
		Compress:   true, // comprime archivos viejos con gzip
	}

	return zapcore.NewCore(
		zapcore.NewJSONEncoder(enc),
		zapcore.AddSync(rotator),
		zapcore.DebugLevel,
	)
}

// errorFileCore — archivo separado solo para errores
// útil para monitorear problemas sin filtrar entre miles de logs de info
func errorFileCore(enc zapcore.EncoderConfig, logDir string) zapcore.Core {
	fileName := filepath.Join(logDir, "error-"+dailyStamp()+".log")

	rotator := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    50,
		MaxBackups: 30,
		MaxAge:     30,
		Compress:   true,
	}

	return zapcore.NewCore(
		zapcore.NewJSONEncoder(enc),
		zapcore.AddSync(rotator),
		zapcore.ErrorLevel, // solo Error y Fatal
	)
}

// dailyStamp devuelve la fecha actual como string para el nombre del archivo
func dailyStamp() string {
	return time.Now().Format("2006-01-02")
}
