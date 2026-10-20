# Examples

Run an example with:

```bash
TYPESAFE_API_KEY=your-api-key go run ./examples/raw
TYPESAFE_API_KEY=your-api-key go run ./examples/builder
TYPESAFE_API_KEY=your-api-key go run ./examples/images ./kitchen.jpg
```

The `raw` example constructs an `EvaluationRequest` directly. The `builder`
example uses the fluent request builder.
The `images` example uploads the supplied image and asks for a Box answer using
Jev 1.14.
