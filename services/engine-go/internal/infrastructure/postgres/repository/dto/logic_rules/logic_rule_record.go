package logic_rules

type IterAlgorithmType string

const (
	LinearAlgorithm IterAlgorithmType = "linear"
)

const (
	JumpAction   = "JMP"
	FinishAction = "FINISH"
)

type LogicRuleRecord struct {
	Action string  `db:"action"`
	Next   *string `db:"next"`
}

type LogicRulesStorageRecord struct {
	Rules          map[string]LogicRuleRecord `json:"rules"`
	DefaultNextAlg IterAlgorithmType          `json:"default_next"`
}
