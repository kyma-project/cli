package nice

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/reconciler/pkg/model"
	"github.com/kyma-incubator/reconciler/pkg/scheduler/service"
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

func (sum *Summary) PrintFailedComponentSummary(result *service.ReconciliationResult) {
	nicePrint := Nice{
		NonInteractive: sum.NonInteractive,
	}
	failedComps := []string{}
	successfulComps := []string{}

	ops := result.GetOperations()

	for _, comp := range ops {
		if comp.State == model.OperationStateError {
			failedComps = append(failedComps, comp.Component)
		}
		if comp.State == model.OperationStateDone {
			successfulComps = append(successfulComps, comp.Component)
		}
	}

	fmt.Println()
	if len(ops) > 0 && ops[0].Type == model.OperationTypeDelete {
		fmt.Printf("Deleted Components: ")
		nicePrint.PrintImportantf("%d/%d", len(successfulComps), len(successfulComps)+len(failedComps))
		fmt.Println("Could not delete the following components:")
	} else {
		fmt.Printf("Deployed Components: ")
		nicePrint.PrintImportantf("%d/%d", len(successfulComps), len(successfulComps)+len(failedComps))
		fmt.Println("Could not deploy the following components:")
	}

	for _, items := range failedComps {
		fmt.Printf("- %s\n", items)
	}
	fmt.Println()
}

func (sum *Summary) Print(t time.Duration) error {
	nicePrint := Nice{
		NonInteractive: sum.NonInteractive,
	}

	// Installation info

	fmt.Println()
	nicePrint.PrintKyma()
	fmt.Print(" is installed in version:\t")
	nicePrint.PrintImportant(sum.Version)

	nicePrint.PrintKyma()
	fmt.Print(" installation took:\t\t")
	nicePrint.PrintImportantf("%s", t.Round(time.Second).String())

	if sum.URL != "" {
		nicePrint.PrintKyma()
		fmt.Print(" is running at:\t\t")
		nicePrint.PrintImportant(sum.URL)
	}

	// Console

	if sum.Console != "" {
		nicePrint.PrintKyma()
		fmt.Print(" console:\t\t\t")
		nicePrint.PrintImportantf(sum.Console)
	}

	if sum.Dashboard != "" {
		nicePrint.PrintKyma()
		fmt.Print(" dashboard:\t\t\t")
		nicePrint.PrintImportantf(sum.Dashboard)
	}

	// Admin credentials

	if sum.Email != "" {
		nicePrint.PrintKyma()
		fmt.Print(" admin email:\t\t")
		nicePrint.PrintImportant(sum.Email)
	}

	fmt.Printf("\nHappy ")
	nicePrint.PrintKyma()
	fmt.Printf("-ing! :)\n\n")

	// TODO refactor function when old deploy goes away, no need to return an error
	return nil
}
