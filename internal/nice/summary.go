package nice

import (
	"fmt"
	"time"
)

type Summary struct {
	NonInteractive bool
	Duration       time.Duration
	Version        string
	URL            string
	Console        string
	Email          string
	Password       string
}

func (s *Summary) Print() error {
	nicePrint := Nice{
		NonInteractive: s.NonInteractive,
	}

	// Installation info

	fmt.Println()
	nicePrint.PrintKyma()
	fmt.Print(" is installed in version:\t")
	nicePrint.PrintImportant(s.Version)

	nicePrint.PrintKyma()
	fmt.Print(" installation took:\t\t")
	nicePrint.PrintImportantf("%d hours %d minutes", int64(s.Duration.Hours()), int64(s.Duration.Minutes()))

	nicePrint.PrintKyma()
	fmt.Print(" is running at:\t\t")
	nicePrint.PrintImportant(s.URL)

	// Console

	nicePrint.PrintKyma()
	fmt.Print(" console:\t\t\t")
	nicePrint.PrintImportantf(s.Console)

	// Admin credentials

	if s.Email != "" {
		nicePrint.PrintKyma()
		fmt.Print(" admin email:\t\t")
		nicePrint.PrintImportant(s.Email)
	}

	if !s.NonInteractive && s.Password != "" {
		nicePrint.PrintKyma()
		fmt.Printf(" admin password:\t\t")
		nicePrint.PrintImportant(s.Password)
	}

	fmt.Printf("\nHappy ")
	nicePrint.PrintKyma()
	fmt.Printf("-ing! :)\n\n")

	return nil
}
