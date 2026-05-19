package control

import "time"

type ConfigEntry struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Scope     string    `json:"scope"`
	Version   int64     `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ABTest struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	FeatureKey     string    `json:"feature_key"`
	Variants       []string  `json:"variants"`
	TrafficPercent int       `json:"traffic_percent"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type Rollout struct {
	ID            string    `json:"id"`
	FeatureKey    string    `json:"feature_key"`
	TargetPercent int       `json:"target_percent"`
	Strategy      string    `json:"strategy"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type ServiceState struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Version   string    `json:"version,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NodeLoad struct {
	NodeID     string    `json:"node_id"`
	Service    string    `json:"service"`
	Status     string    `json:"status"`
	CPUPercent int       `json:"cpu_percent"`
	MemPercent int       `json:"mem_percent"`
	Rooms      int       `json:"rooms"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type UpdatePlan struct {
	ID          string    `json:"id"`
	Service     string    `json:"service"`
	Version     string    `json:"version"`
	Strategy    string    `json:"strategy"`
	Status      string    `json:"status"`
	RequestedAt time.Time `json:"requested_at"`
}
