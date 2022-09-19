package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

type Input struct {
	Query  string      `json:"query"`
	Params interface{} `json:"params"`
}

func main() {
	var err error

	scriptPath := "/Users/a.lacoin/Developer/pomdtr/sunbeam/scripts/github.mjs"
	cmd := exec.Command(scriptPath)

	var inbuf, outbuf, errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	cmd.Stdout = &outbuf
	cmd.Stdin = &inbuf

	var bytes []byte
	bytes, err = json.Marshal(Input{})
	inbuf.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Run()

	if err != nil {
		log.Fatalf("%s: %s", err, errbuf.String())
	}

	fmt.Println(outbuf.String())
}
