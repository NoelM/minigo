package logs

import (
	"log"
	"os"
)

var infoLog = log.New(os.Stdout, "[notel] info:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var warnLog = log.New(os.Stdout, "[notel] warn:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var errorLog = log.New(os.Stdout, "[notel] error:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

func InfoLog(format string, v ...any) {
	infoLog.Printf(format, v...)
}

func WarnLog(format string, v ...any) {
	warnLog.Printf(format, v...)
}

func ErrorLog(format string, v ...any) {
	errorLog.Printf(format, v...)
}
