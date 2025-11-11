package main

import (
	"gadmin"
	"os"
	"strings"
)

func main() {
	cwd, _ := os.Getwd()
	if strings.HasSuffix(cwd, "cmd") {
		_ = os.Chdir("..")
	}

	admin := gadmin.NewAdmin("Admin", nil)

	// if db, err := gadmin.Parse(dao.Url()).OpenDefault(); err == nil {
	// 	admin.DB = db
	// }

	// for _, v := range dao.Views() {
	// 	admin.AddView(v)
	// }

	admin.Run()
}
