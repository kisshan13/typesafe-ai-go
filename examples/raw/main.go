package main

import (
	"context"
	"fmt"
	"os"

	typesafe "github.com/kisshan13/typesafe-ai-go"
)

func main() {
	client := typesafe.NewClient(os.Getenv("TYPESAFE_API_KEY"))
	request := typesafe.EvaluationRequest{
		State: "Help! My payouts have been failing for 3 days.",
		Model: typesafe.JevLatest,
		Questions: map[string]typesafe.Question{
			"is_urgent": typesafe.Noul("Does this convey urgency?", &typesafe.NoulCriteria{
				True:  "Explicitly time-sensitive",
				False: "No urgency expressed",
			}),
			"department": typesafe.Choice("Which team should handle this?", map[string]any{
				"billing":   "Payments, invoicing, refunds",
				"technical": "Bugs, outages, integrations",
			}),
		},
	}

	result, err := client.Evaluate(context.Background(), request)
	if err != nil {
		panic(err)
	}

	department, ok := result.Answers.Get("department")
	if ok {
		fmt.Println(department.Choice)
	}
}
