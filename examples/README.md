# Examples

Run an example with:

```bash
TYPESAFE_API_KEY=your-api-key go run ./examples/raw
TYPESAFE_API_KEY=your-api-key go run ./examples/builder
```

The `raw` example constructs an `EvaluationRequest` directly. The `builder`
example uses the fluent request builder.
