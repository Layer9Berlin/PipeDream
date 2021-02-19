package graph

import (
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
)

var openInBrowser = func(file string) error {
	openCommand := exec.Command("open", file)
	return openCommand.Run()
}

func OutputGraph(executionContext interface{}) {
	ut, err := template.New("runs").Parse(graphTemplate())

	if err != nil {
		panic(err)
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), "*.html")
	if err != nil {
		panic(err)
	}
	err = ut.Execute(tmpFile, executionContext)

	if err != nil {
		panic(err)
	}

	err = tmpFile.Close()
	if err != nil {
		panic(err)
	}

	err = openInBrowser(tmpFile.Name())
	if err != nil {
		panic(err)
	}
}
