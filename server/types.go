package server

import (
	"github.com/infor-design/selfservice/pkg/client"
	"github.com/infor-design/selfservice/pkg/db"
	"github.com/infor-design/selfservice/reposerver"
	v1 "k8s.io/api/batch/v1"
)

type RepoResp struct {
	App       db.Application                `json:"app"`
	Manifests *reposerver.ManifestsResponse `json:"manifests"`
}

type SettingsHttpResponse struct {
	RepoRoot   string `json:"repo_root"`
	SshRoot    string `json:"ssh_root"`
	PrivateKey string `json:"private_key"`
}

type AppManifestHttpResp struct {
	App       db.Application                `json:"app"`
	Manifests *reposerver.ManifestsResponse `json:"manifests"`
}

type JobRunResponse struct {
	Job    db.Job           `json:"job"`
	Config client.JobConfig `json:"config"`
	Spec   v1.JobSpec       `json:"spec"`
	Status v1.JobStatus     `json:"status"`
}
