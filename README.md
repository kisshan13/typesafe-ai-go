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
base URL is `https://api.typesafe.ai`; model names use the `typesafe.Model`
string type. Configure another endpoint with
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

## Images and Box questions

Use `ImageFromFile` or `NewImage` to place an image in the state, instructions,
or criteria. The client creates image references, uploads the files as
`multipart/form-data`, and adds the required images header automatically. Image
and Box support is experimental. Box questions require Jev 1.14 (`JevPreview`
or `Jev114`).

```go
photo, err := typesafe.ImageFromFile("kitchen.jpg")
if err != nil {
	panic(err)
}

request := typesafe.NewEvaluationRequest(map[string]any{"photo": photo}).
	Model(typesafe.JevPreview).
	Box("mug", "the coffee mug in `photo`").
	Noul("mug_visible", "Is there a coffee mug in `photo`?", nil).
	Build()

result, err := client.Evaluate(ctx, request)
if err != nil {
	panic(err)
}
edges := result.Answers.MustGet("mug").Box
fmt.Println(edges.Left, edges.Top, edges.Right, edges.Bottom)
```

Image bytes may also be constructed directly with `typesafe.NewImage(data,
"jpg")`. The `Answer.Box` field contains normalized `Top`, `Bottom`, `Left`,
and `Right` edges. To place an image inside a longer instruction string, add it
with a reference using `.Image("reference", photo)` and include that reference
in the string; the client uploads the image for you.

The client retries `429 Too Many Requests` and `529 Overloaded` responses with
exponential backoff. Use `WithBaseURL`, `WithHTTPClient`, `WithMaxRetries`, and
`WithRetryBackoff` to customize it.

## License

This project is available under the [MIT License](./LICENSE).
