package typesafe

const DefaultBaseURL = "https://api.typesafe.ai"

type Model string

const JevLatest Model = "jev-latest"
const JevPreview Model = "jev-preview"
const Jev114 Model = "jev-1.14.0"
const DefaultModel = JevLatest

type Question = any

type EvaluationRequest struct {
	State     any                 `json:"state"`
	Model     Model               `json:"model"`
	Questions map[string]Question `json:"questions"`
	Assets    map[string]Asset    `json:"assets,omitempty"`
}

type Asset struct {
	Type       string `json:"type,omitempty"`
	Part       string `json:"part,omitempty"`
	DataBase64 string `json:"data_base64,omitempty"`
	Format     string `json:"format,omitempty"`
	Data       []byte `json:"-"`
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

type BoxQuestion struct {
	Type         string `json:"type"`
	Instructions any    `json:"instructions"`
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

func Box(instructions any) BoxQuestion {
	return BoxQuestion{Type: "box", Instructions: instructions}
}

type EvaluationRequestBuilder struct {
	state     any
	model     Model
	questions map[string]Question
	assets    map[string]Asset
}

func NewEvaluationRequest(state any) *EvaluationRequestBuilder {
	return &EvaluationRequestBuilder{
		state:     state,
		model:     JevLatest,
		questions: make(map[string]Question),
	}
}

func (b *EvaluationRequestBuilder) Model(model Model) *EvaluationRequestBuilder {
	if model != "" {
		b.model = model
	}
	return b
}

func (b *EvaluationRequestBuilder) Image(reference string, image Image) *EvaluationRequestBuilder {
	if b.assets == nil {
		b.assets = make(map[string]Asset)
	}
	b.assets[reference] = Asset{Type: "image", Format: image.format, Data: append([]byte(nil), image.data...)}
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

func (b *EvaluationRequestBuilder) Box(id string, instructions any) *EvaluationRequestBuilder {
	return b.Question(id, Box(instructions))
}

func (b *EvaluationRequestBuilder) Build() EvaluationRequest {
	return EvaluationRequest{State: b.state, Model: b.model, Questions: b.questions, Assets: b.assets}
}

type EvaluationResponse struct {
	Model   Model   `json:"model"`
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
	Box           *BoundingBox       `json:"box,omitempty"`
}

type BoundingBox struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}
