package workspace

import "time"

type Status string

const (
	StatusQueued       Status = "queued"
	StatusProvisioning Status = "provisioning"
	StatusIdle         Status = "idle"
	StatusBusy         Status = "busy"
	StatusError        Status = "error"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

type Workspace struct {
	ID        string    `dynamodbav:"id"`
	Channel   string    `dynamodbav:"channel"`
	ThreadTS  string    `dynamodbav:"threadTs"`
	Status    Status    `dynamodbav:"status"`
	RootUser  string    `dynamodbav:"rootUser"`
	RootText  string    `dynamodbav:"rootText"`
	VMID      string    `dynamodbav:"vmId"`
	CreatedAt time.Time `dynamodbav:"createdAt"`
	UpdatedAt time.Time `dynamodbav:"updatedAt"`
}

type Message struct {
	Role        Role      `dynamodbav:"role"`
	Author      string    `dynamodbav:"author"`
	Text        string    `dynamodbav:"text"`
	SlackTS     string    `dynamodbav:"slackTs"`
	ClientMsgID string    `dynamodbav:"clientMsgId"`
	CreatedAt   time.Time `dynamodbav:"createdAt"`
}
