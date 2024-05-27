package util

import (
	"log"
	"os"
)

var Log = log.New(os.Stdout, "[cheryl]", log.Lshortfile|log.Ldate|log.Ltime)