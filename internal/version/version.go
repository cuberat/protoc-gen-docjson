package version

import (
	"fmt"
	"os"
	"path"
)

const version string = "0.1"

func GetVersion() string {
	return version
}

func PrintVersion() {
	fmt.Printf("%s %s\n", path.Base(os.Args[0]), GetVersion())
}
