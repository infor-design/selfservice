package job

import (
	"github.com/infor-design/selfservice/pkg/db"
)

type Job struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	ApplicationID uint   `json:"application_id"`
	Phase         string `json:"phase"`
	Spec          string `json:"spec"`
	Created_At    string `json:"created_at"`
	Updated_At    string `json:"updated_at"`
	Deleted_At    string `json:"deleted_at"`
}

type JobService struct {
	db *db.Connection
}
