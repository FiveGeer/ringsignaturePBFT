package mylog

import (
	"os"
	"log"
)

var (
	Info *log.Logger
	Warning *log.Logger
	Error * log.Logger
)

func init(){
	Info = log.New(os.Stdout,"Info:",log.Ldate | log.Ltime | log.Lshortfile)
	Warning = log.New(os.Stdout,"Warning:",log.Ldate | log.Ltime | log.Lshortfile)
	Error = log.New(os.Stderr,"Error:",log.Ldate | log.Ltime | log.Lshortfile)
}

