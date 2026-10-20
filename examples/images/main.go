package main

import (
	"context"
	"fmt"
	"os"

	typesafe "github.com/kisshan13/typesafe-ai-go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run . <image-file>")
		os.Exit(2)
	}

	photo, err := typesafe.ImageFromFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	client := typesafe.NewClient(os.Getenv("TYPESAFE_API_KEY"))
	request := typesafe.NewEvaluationRequest(map[string]any{"photo": photo}).
		Model(typesafe.JevPreview).
		Box("mug", "the coffee mug in `photo`").
		Noul("mug_visible", "Is there a coffee mug in `photo`?", nil).
		Build()

	result, err := client.Evaluate(context.Background(), request)
	if err != nil {
		panic(err)
	}

	box := result.Answers.MustGet("mug").Box
	fmt.Printf("top=%.3f bottom=%.3f left=%.3f right=%.3f\n", box.Top, box.Bottom, box.Left, box.Right)
	fmt.Printf("mug visible: %.3f\n", result.Answers.MustGet("mug_visible").Noul)
}
