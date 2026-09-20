package main

import (
	"context"
	"fmt"
	"os"

	typesafe "github.com/kisshan13/typesafe-ai-go"
)

func main() {
	client := typesafe.NewClient(os.Getenv("TYPESAFE_API_KEY"))
	request := typesafe.NewEvaluationRequest("Help! My payouts have been failing for 3 days.").
		Noul("is_urgent", "Does this convey urgency?", &typesafe.NoulCriteria{
			True:  "Explicitly time-sensitive",
			False: "No urgency expressed",
		}).
		Choice("department", "Which team should handle this?", map[string]any{
			"billing":   "Payments, invoicing, refunds",
			"technical": "Bugs, outages, integrations",
		}).
		Score("frustration", "How frustrated is the customer?", []any{
			"Calm", "Frustrated", "Very angry",
		}).
		Build()

	result, err := client.Evaluate(context.Background(), request)
	if err != nil {
		panic(err)
	}

	department := result.Answers.MustGet("department")
	fmt.Println(department.Choice)
}
