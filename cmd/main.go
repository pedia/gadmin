package main

import (
	"gadm"
	"os"
	"strings"
)

func main() {
	cwd, _ := os.Getwd()
	if strings.HasSuffix(cwd, "cmd") {
		_ = os.Chdir("..")
	}

	gadm.NewAdmin("Admin").Run()
}
