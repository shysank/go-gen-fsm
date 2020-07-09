package main

import (
	"bytes"
	"fmt"
	go_gen_fsm "github.com/cynic89/go-gen-fsm"
)

type Door struct {
	code  string
	sofar bytes.Buffer
}

func (d *Door) Init(args ...interface{}) string {
	arg := args[0].([]interface{})
	d.code = arg[0].(string)
	return "Locked"
}

func (d *Door) Locked_button(digit rune) string {
	d.sofar.WriteRune(digit)
	sofarStr := d.sofar.String()
	fmt.Println(sofarStr)
	if sofarStr == d.code {
		fmt.Println("Opened")
		return "Open"
	}

	if len(sofarStr) < len(d.code) {
		fmt.Println("parital unlock")
		return "Locked"
	}

	fmt.Println("Wrong code")

	d.sofar.Reset()
	return "Locked"

}

func (d *Door) Open_timeout() string {
	fmt.Println("timeout, going back to locked")
	d.sofar.Reset()
	return "Locked"
}

func main() {
	d := &Door{}

	ch, clo := go_gen_fsm.Start(d, "pass")

	for {
		go_gen_fsm.SendEvent(ch, "button", 'p')
		go_gen_fsm.SendEvent(ch, "button", 'a')
		go_gen_fsm.SendEvent(ch, "button", 's')
		go_gen_fsm.SendEvent(ch, "button", 's')

		<-clo
	}

}
