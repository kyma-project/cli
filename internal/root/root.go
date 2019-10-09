package root

import "fmt"

func PromptUser() bool {
	for {
		fmt.Print("Type [Y/n]: ")
		var res string
		if _, err := fmt.Scanf("%s", &res); err != nil {
			return false
		}
		switch res {
		case "Yes", "Y":
			return true
		case "no", "n":
			return false
		default:
			continue
		}
	}
}
