package fix

import (
	"log"
	"os"
)

func init() {
	var err error
	if _, err = os.Stat("_test/"); os.IsNotExist(err) {
		err = os.Mkdir("_test/", 0755)
	}

	if err != nil {
		log.Fatal(err)
	}
}
