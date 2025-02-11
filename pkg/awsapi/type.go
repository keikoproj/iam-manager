package awsapi

// PolicyDocument represents the structure of an IAM policy document
type PolicyDocument struct {
	Statement []StatementEntry `json:"Statement"`
}

// StatementEntry represents a single statement in an IAM policy document
type StatementEntry struct {
	Action   interface{} `json:"Action"`
	Resource interface{} `json:"Resource"`
	Effect   string      `json:"Effect"`
}
