package logging

import (
	"log"
	"os"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
)

func SetUpLogger(logFilePath string) {
	var output *os.File = os.Stdout

	if logFilePath != "" {
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Println("Could not open log file:", err, "- falling back to stdout")
		} else {
			output = file
		}
	}

	infoLogger = log.New(output, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(output, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLogger = log.New(output, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(v ...any) {
	infoLogger.Println(v...)
}

func Error(v ...any) {
	errorLogger.Println(v...)
}

func Debug(v ...any) {
	debugLogger.Println(v...)
}
