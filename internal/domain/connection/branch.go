package connection

import "errors"

type ConditionBranch string

const (
	ConditionBranchTrue  ConditionBranch = "true"
	ConditionBranchFalse ConditionBranch = "false"
)

func (b ConditionBranch) Valid() bool {
	switch b {
	case ConditionBranchTrue, ConditionBranchFalse:
		return true
	default:
		return false
	}
}

func ParseConditionBranch(value string) (ConditionBranch, error) {
	b := ConditionBranch(value)
	if !b.Valid() {
		return "", errors.New("invalid condition branch")
	}
	return b, nil
}
