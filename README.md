# TypeSafe Go SDK

Go client for the TypeSafe System One evaluation API.

This is not an official TypeSafe AI Go SDK. It is a community-maintained
project primarily maintained by [Kisshan13](https://github.com/kisshan13).

## Installation

Install the package with:

```bash
go get github.com/kisshan13/typesafe-ai-go
```

Then import it in your Go program:

```go
import "github.com/kisshan13/typesafe-ai-go"
```

The default model is `typesafe.JevLatest` (`"jev-latest"`). You can omit
`EvaluationRequest.Model` to use it, or set it explicitly. The default API
base URL is `https://api.typesafe.ai`; configure another endpoint with
`typesafe.WithBaseURL("https://example.com")`. Passing an empty URL restores
the default endpoint.

## Raw request

The raw pattern gives you direct control over the request structs and question map.

```go
client := typesafe.NewClient(os.Getenv("TYPESAFE_API_KEY"))

result, err := client.Evaluate(ctx, typesafe.EvaluationRequest{
	State: "Help! My payouts have been failing for 3 days.",
	Model: typesafe.JevLatest,
	Questions: map[string]typesafe.Question{
		"is_urgent": typesafe.Noul("Does this convey urgency?", &typesafe.NoulCriteria{
			True: "Explicitly time-sensitive",
			False: "No urgency expressed",
		}),
		"department": typesafe.Choice("Which team should handle this?", map[string]any{
			"billing":  "Payments, invoicing, refunds",
			"technical": "Bugs, outages, integrations",
		}),
		"frustration": typesafe.Score("How frustrated is the customer?", []any{
			"Calm", "Frustrated", "Very angry",
		}),
	},
})
if err != nil {
	panic(err)
}
department, ok := result.Answers.Get("department")
if ok {
	fmt.Println(department.Choice)
}
```

## Builder request

The builder provides a concise developer-friendly pattern while producing the
same `EvaluationRequest` used by `Client.Evaluate`.

```go
request := typesafe.NewEvaluationRequest("Help! My payouts have been failing for 3 days.").
	Noul("is_urgent", "Does this convey urgency?", &typesafe.NoulCriteria{
		True: "Explicitly time-sensitive",
		False: "No urgency expressed",
	}).
	Choice("department", "Which team should handle this?", map[string]any{
		"billing":  "Payments, invoicing, refunds",
		"technical": "Bugs, outages, integrations",
	}).
	Score("frustration", "How frustrated is the customer?", []any{
		"Calm", "Frustrated", "Very angry",
	}).
	Build()

result, err := client.Evaluate(ctx, request)
if err != nil {
	panic(err)
}

department := result.Answers.MustGet("department")
fmt.Println(department.Choice)
```

`Answers.Get` returns `(Answer, bool)` for safe lookups. `Answers.MustGet`
returns the answer directly and panics when the requested ID is missing.

Additional runnable examples are available in the [examples](./examples)
directory.

The client retries `429 Too Many Requests` and `529 Overloaded` responses with
exponential backoff. Use `WithBaseURL`, `WithHTTPClient`, `WithMaxRetries`, and
`WithRetryBackoff` to customize it.

## License

This project is available under the [MIT License](./LICENSE).
