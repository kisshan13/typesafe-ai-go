package typesafe

const DefaultBaseURL = "https://api.typesafe.ai"
const JevLatest = "jev-latest"
const DefaultModel = JevLatest

type Question = any

type EvaluationRequest struct {
	State     any                 `json:"state"`
	Model     string              `json:"model"`
	Questions map[string]Question `json:"questions"`
}

type NoulQuestion struct {
	Type         string `json:"type"`
	Instructions any    `json:"instructions"`
	Criteria     any    `json:"criteria,omitempty"`
}

type NoulCriteria struct {
	True  any `json:"true,omitempty"`
	False any `json:"false,omitempty"`
}

type ChoiceQuestion struct {
	Type         string         `json:"type"`
	Instructions any            `json:"instructions"`
	Criteria     map[string]any `json:"criteria"`
}

type ScoreQuestion struct {
	Type         string `json:"type"`
	Instructions any    `json:"instructions"`
	Criteria     []any  `json:"criteria"`
}

func Noul(instructions any, criteria any) NoulQuestion {
	return NoulQuestion{Type: "noul", Instructions: instructions, Criteria: criteria}
}

func Choice(instructions any, criteria map[string]any) ChoiceQuestion {
	return ChoiceQuestion{Type: "choice", Instructions: instructions, Criteria: criteria}
}

func Score(instructions any, criteria []any) ScoreQuestion {
	return ScoreQuestion{Type: "score", Instructions: instructions, Criteria: criteria}
}

type EvaluationRequestBuilder struct {
	state     any
	model     string
	questions map[string]Question
}

func NewEvaluationRequest(state any) *EvaluationRequestBuilder {
	return &EvaluationRequestBuilder{
		state:     state,
		model:     JevLatest,
		questions: make(map[string]Question),
	}
}

func (b *EvaluationRequestBuilder) Model(model string) *EvaluationRequestBuilder {
	if model != "" {
		b.model = model
	}
	return b
}

func (b *EvaluationRequestBuilder) Question(id string, question Question) *EvaluationRequestBuilder {
	b.questions[id] = question
	return b
}

func (b *EvaluationRequestBuilder) Noul(id string, instructions any, criteria any) *EvaluationRequestBuilder {
	return b.Question(id, Noul(instructions, criteria))
}

func (b *EvaluationRequestBuilder) Choice(id string, instructions any, criteria map[string]any) *EvaluationRequestBuilder {
	return b.Question(id, Choice(instructions, criteria))
}

func (b *EvaluationRequestBuilder) Score(id string, instructions any, criteria []any) *EvaluationRequestBuilder {
	return b.Question(id, Score(instructions, criteria))
}

func (b *EvaluationRequestBuilder) Build() EvaluationRequest {
	return EvaluationRequest{State: b.state, Model: b.model, Questions: b.questions}
}

type EvaluationResponse struct {
	Model   string  `json:"model"`
	Answers Answers `json:"answers"`
	Usage   Usage   `json:"usage"`
}

type Answers map[string]Answer

func (a Answers) Get(id string) (Answer, bool) {
	answer, ok := a[id]
	return answer, ok
}

func (a Answers) MustGet(id string) Answer {
	answer, ok := a.Get(id)
	if !ok {
		panic("typesafe: answer not found: " + id)
	}
	return answer
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type Answer struct {
	Type          string             `json:"type"`
	Noul          float64            `json:"noul,omitempty"`
	Choice        string             `json:"choice,omitempty"`
	Score         float64            `json:"score,omitempty"`
	Legend        map[string]string  `json:"legend,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Confidence    float64            `json:"confidence,omitempty"`
}
