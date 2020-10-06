package root

import "fmt"

func PromptUser() bool {
	for {
		fmt.Print("Type [y/N]: ")
		var res string
		if _, err := fmt.Scanf("%s", &res); err != nil {
			return false
		}
		switch res {
		case "yes", "y":
			return true
		case "No", "N", "no", "n":
			return false
		default:
			continue
		}
	}
}
