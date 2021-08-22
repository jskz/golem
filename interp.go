package main

type Command struct {
	Name    string
	CmdFunc func(ch *Character, arguments string)
}

var CommandTable map[string]Command

func InitCommandTable() {
	CommandTable = make(map[string]Command)

	CommandTable["score"] = Command{Name: "score", CmdFunc: do_score}
}

func (ch *Character) Interpret(input string) {

}
