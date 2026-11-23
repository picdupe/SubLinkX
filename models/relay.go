package models

import (
	"time"

	"gorm.io/gorm"
)

// RelayServer stores the control-plane identity and observed state of a relay host.
type RelayServer struct {
	gorm.Model
	Name                string     `gorm:"size:100;not null" json:"name"`
	PublicHost          string     `gorm:"size:255;not null" json:"public_host"`
	AgentID             string     `gorm:"size:64;not null;uniqueIndex" json:"agent_id"`
	AgentTokenHash      string     `gorm:"size:64;not null" json:"-"`
	AgentTokenCreatedAt *time.Time `json:"agent_token_created_at"`
	AgentVersion        string     `gorm:"size:32" json:"agent_version"`
	OSInfo              string     `gorm:"size:255" json:"os_info"`
	Architecture        string     `gorm:"size:32" json:"architecture"`
	ServiceManager      string     `gorm:"size:20" json:"service_manager"`
	HAProxyVersion      string     `gorm:"size:64" json:"haproxy_version"`
	HAProxyStatus       string     `gorm:"size:20;not null;default:unknown" json:"haproxy_status"`
	DesiredRevision     uint64     `gorm:"not null;default:0" json:"desired_revision"`
	AppliedRevision     uint64     `gorm:"not null;default:0" json:"applied_revision"`
	HeartbeatAt         *time.Time `json:"last_heartbeat_at"`
	Enabled             bool       `gorm:"not null;default:true;index" json:"enabled"`
}

// RelayRoute connects a public relay listener to an existing SublinkX node.
type RelayRoute struct {
	gorm.Model
	Subscriptions  []Subcription `gorm:"many2many:subcription_relay_routes;" json:"-"`
	Name           string        `gorm:"size:100;not null" json:"name"`
	RelayServerID  uint          `gorm:"not null;uniqueIndex:ux_relay_server_protocol_port;index" json:"relay_server_id"`
	LandingNodeID  int           `gorm:"not null;index" json:"landing_node_id"`
	Protocol       string        `gorm:"size:10;not null;uniqueIndex:ux_relay_server_protocol_port;default:tcp" json:"protocol"`
	ListenAddress  string        `gorm:"size:64;not null;default:0.0.0.0" json:"listen_address"`
	ListenPort     uint16        `gorm:"not null;uniqueIndex:ux_relay_server_protocol_port" json:"listen_port"`
	Enabled        bool          `gorm:"not null;default:true;index" json:"enabled"`
	PublishEnabled bool          `gorm:"not null;default:false;index" json:"publish_enabled"`
	Priority       int           `gorm:"not null;default:0" json:"priority"`
	Description    string        `gorm:"type:text" json:"description"`
}

type RelayConfigRevision struct {
	gorm.Model
	RelayServerID uint   `gorm:"not null;uniqueIndex:ux_relay_server_revision" json:"relay_server_id"`
	Revision      uint64 `gorm:"not null;uniqueIndex:ux_relay_server_revision" json:"revision"`
	ConfigSHA256  string `gorm:"size:64;not null" json:"config_sha256"`
	ConfigContent string `gorm:"type:text;not null" json:"config_content"`
	ChangeSummary string `gorm:"type:text" json:"change_summary"`
	CreatedBy     uint   `gorm:"not null;default:0" json:"created_by"`
}

type RelayConfigTask struct {
	gorm.Model
	TaskID         string     `gorm:"size:36;not null;uniqueIndex" json:"task_id"`
	RelayServerID  uint       `gorm:"not null;index" json:"relay_server_id"`
	RevisionID     uint       `gorm:"not null;index" json:"revision_id"`
	TaskType       string     `gorm:"size:32;not null" json:"task_type"`
	Status         string     `gorm:"size:20;not null;index" json:"status"`
	Attempt        uint       `gorm:"not null;default:0" json:"attempt"`
	LeaseExpiresAt *time.Time `json:"lease_expires_at"`
	ErrorCode      string     `gorm:"size:64" json:"error_code"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
}

type RelayAgentEvent struct {
	gorm.Model
	RelayServerID uint   `gorm:"not null;index" json:"relay_server_id"`
	EventType     string `gorm:"size:64;not null;index" json:"event_type"`
	TaskID        string `gorm:"size:36;index" json:"task_id,omitempty"`
	Details       string `gorm:"type:text" json:"details,omitempty"`
}
