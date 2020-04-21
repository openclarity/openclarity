package config

type DeleteJobPolicyType string

const (
	DeleteJobPolicyAll        DeleteJobPolicyType = "All"
	DeleteJobPolicyNever      DeleteJobPolicyType = "Never"
	DeleteJobPolicySuccessful DeleteJobPolicyType = "Successful"
)

func (dj DeleteJobPolicyType) IsValid() bool {
	switch dj {
	case DeleteJobPolicyAll, DeleteJobPolicyNever, DeleteJobPolicySuccessful:
		return true
	default:
		return false
	}
}
