package model

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ApiKey struct {
	ID        int64  `json:"id"`
	Label     string `json:"label"`
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name"`
	CreatedAt string `json:"created_at"`
}

type Task struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	Tags        []string   `json:"tags,omitempty"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

type TaskStatus string

const (
	StatusBacklog    TaskStatus = "backlog"
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in-progress"
	StatusReview     TaskStatus = "review"
	StatusDone       TaskStatus = "done"
)

type TaskFilter struct {
	ProjectID *string
	Status    *TaskStatus
	Limit     int
	Offset    int
}

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Active    bool   `json:"active"`
	IsAdmin   bool   `json:"is_admin"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s TaskStatus) String() string {
	return string(s)
}

func ValidStatus(s string) bool {
	switch TaskStatus(s) {
	case StatusBacklog, StatusTodo, StatusInProgress, StatusReview, StatusDone:
		return true
	default:
		return false
	}
}
