package nice

import (
	"fmt"
	"github.com/kyma-incubator/reconciler/pkg/model"
	"github.com/kyma-incubator/reconciler/pkg/scheduler/service"
	"time"
)

type Summary struct {
	NonInteractive bool
	Version        string
	URL            string
	Console        string
	Dashboard      string
	Email          string
	Password       string
}

func (s *Summary) PrintFailedComponentSummary(result *service.ReconciliationResult) error{
	nicePrint := Nice{
		NonInteractive: s.NonInteractive,
	}
	failedComps := []string{}
	successfulComps := []string{}

	for _, comp := range result.GetOperations() {
		if comp.State == model.OperationStateError {
			failedComps = append(failedComps, comp.Component)
		}
		if comp.State == model.OperationStateDone {
			successfulComps = append(successfulComps, comp.Component)
		}
	}

	fmt.Println()
	fmt.Printf("Deployed Components: ")
	nicePrint.PrintImportantf("%d/%d", len(successfulComps), len(successfulComps) + len(failedComps))
	fmt.Println("Failed:")
	for _, items := range failedComps {
		fmt.Printf("- %s\n", items)
	}
	fmt.Println()
	return nil

}

func (s *Summary) Print(t time.Duration) error {
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
	nicePrint.PrintImportantf("%d hours %d minutes", int64(t.Hours()), int64(t.Minutes()))

	if s.URL != "" {
		nicePrint.PrintKyma()
		fmt.Print(" is running at:\t\t")
		nicePrint.PrintImportant(s.URL)
	}

	// Console

	if s.Console != "" {
		nicePrint.PrintKyma()
		fmt.Print(" console:\t\t\t")
		nicePrint.PrintImportantf(s.Console)
	}

	if s.Dashboard != "" {
		nicePrint.PrintKyma()
		fmt.Print(" dashboard:\t\t\t")
		nicePrint.PrintImportantf(s.Dashboard)
	}

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


