package internal

type Config struct {
	Workers    int    `json:"workers"`
	RPS        int    `json:"requests_per_second"`
	BatchSize  int    `json:"batch_size"`
	MinID      int    `json:"min_group_id"`
	MaxID      int    `json:"max_group_id"`
	Timeout    string `json:"timeout"`
	WebhookURL string `json:"webhook_url"`
	Unique     bool   `json:"unique"`
}

type Group struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Owner              any    `json:"owner"`
	PublicEntryAllowed bool   `json:"publicEntryAllowed"`
	IsLocked           bool   `json:"isLocked"`
}

type GroupsResponse struct {
	Data []Group `json:"data"`
}

type Stats struct {
	Requests    uint64
	Checked     uint64
	Hits        uint64
	RateLimited uint64
	Errors      uint64
}
