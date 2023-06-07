package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/infor-design/selfservice/pkg/application"
	"github.com/infor-design/selfservice/pkg/client"
	"github.com/infor-design/selfservice/pkg/job"
	repoPkg "github.com/infor-design/selfservice/pkg/repo"
	"github.com/infor-design/selfservice/reposerver"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func reposHandler(service *repoPkg.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		conn, err := grpc.Dial(":9000", grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Errorln(err)
		}

		defer conn.Close()

		rp := reposerver.NewRepoServiceClient(conn)

		switch r.Method {
		case "GET":
			repos := service.List()
			reposBytes, err := json.Marshal(repos)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(reposBytes))
		case "POST":
			var newRepoPayload repoPkg.RepoCreate
			err := decodeJSONBody(rw, r, &newRepoPayload)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			var newRepo repoPkg.Repo
			newRepo.Url = newRepoPayload.Url
			repo := service.Create(newRepo)

			if len(newRepoPayload.Ssh_Private_Key) > 0 {
				message := reposerver.SaveSshKeyRequest{SshKey: newRepoPayload.Ssh_Private_Key, RepoId: strconv.FormatInt(int64(repo.ID), 10)}
				_, err = rp.SaveSshKey(context.Background(), &message)
			}

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			newRepoBytes, err := json.Marshal(repo)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(newRepoBytes))
		default:
			JSONError(rw, errorResp{Message: "Something went wrong..."}, http.StatusInternalServerError)
		}
	}
}

func repoHandler(repoService *repoPkg.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		conn, err := grpc.Dial(":9000", grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Errorln(err)
		}

		defer conn.Close()

		rp := reposerver.NewRepoServiceClient(conn)
		vars := mux.Vars(r)
		repoId := vars["id"]
		idAsUInt, err := strconv.ParseUint(repoId, 10, 32)

		if err != nil {
			log.Errorln(err)
		}

		switch r.Method {
		case "GET":
			repo, err := repoService.Get(uint(idAsUInt))

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			repoBytes, err := json.Marshal(repo)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(repoBytes))
			return
		case "POST":
			repo, err := repoService.Get(uint(idAsUInt))

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			action := vars["action"]

			if action == "sync" {
				message := reposerver.SyncRequest{Repo: repo.Url, RepoId: strconv.FormatInt(int64(repo.ID), 10)}
				resp, err := rp.Sync(context.Background(), &message)

				if err != nil {
					JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
					return
				}

				repo.Commit = resp.Commit
				repo.Hash = resp.Hash
				repoService.Update(repo)
			}

			repoBytes, err := json.Marshal(repo)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(repoBytes))
			return
		case "PUT":
			repo, err := repoService.Get(uint(idAsUInt))

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			var updateRepoPayload repoPkg.RepoUpdate
			err = decodeJSONBody(rw, r, &updateRepoPayload)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			repo.Url = updateRepoPayload.Url
			repoService.Update(repo)

			if len(updateRepoPayload.Ssh_Private_Key) > 0 {
				message := reposerver.SaveSshKeyRequest{SshKey: updateRepoPayload.Ssh_Private_Key, RepoId: strconv.FormatInt(int64(repo.ID), 10)}
				_, err = rp.SaveSshKey(context.Background(), &message)
			}

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			repoBytes, err := json.Marshal(repoPkg.Repo{
				Id:         int(repo.ID),
				Url:        repo.Url,
				Commit:     repo.Commit,
				Hash:       repo.Hash,
				Created_At: repo.CreatedAt.String(),
				Updated_At: repo.UpdatedAt.String(),
				Deleted_At: repo.DeletedAt.Time.String(),
			})

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(repoBytes))
			return
		case "DELETE":
			repo, err := repoService.Get(uint(idAsUInt))

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			err = repoService.Delete(repo)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			message := reposerver.RemoveSshKeyRequest{RepoId: strconv.FormatInt(int64(repo.ID), 10)}
			_, err = rp.RemoveSshKey(context.Background(), &message)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			http.Error(rw, "", http.StatusNoContent)
		default:
			JSONError(rw, errorResp{Message: "Something went wrong..."}, http.StatusInternalServerError)
		}
	}
}

func applicationHandler(applicationService *application.Service, repoService *repoPkg.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		conn, err := grpc.Dial(":9000", grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Errorln(err)
		}

		defer conn.Close()

		rp := reposerver.NewRepoServiceClient(conn)

		vars := mux.Vars(r)
		applicationID := vars["id"]
		idAsUInt, err := strconv.ParseUint(applicationID, 10, 32)
		resp := AppManifestHttpResp{}

		if err != nil {
			log.Errorln(err)
		}

		switch r.Method {
		case "GET":
			app, err := applicationService.Get(uint(idAsUInt))

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			repo, err := repoService.Get(app.RepoID)

			if err != nil {
				log.Errorln(err)
			}

			if err == nil {
				repoDirRequest := reposerver.RepoDirRequest{RepoUrl: repo.Url}
				repoDirResponse, err := rp.GetRepoDir(context.Background(), &repoDirRequest)

				if err != nil {
					log.Errorln(err)
				}

				fullManifestPath := path.Join(repoDirResponse.Path, app.ManifestPath)
				message := reposerver.ManifestsRequest{Path: fullManifestPath}
				response, err := rp.GetManifests(context.Background(), &message)

				if err != nil {
					log.Errorln(err)
				}

				resp.Manifests = response
			}

			resp.App = app
			respBytes, err := json.Marshal(resp)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(respBytes))
			return
		case "PUT":
			app, err := applicationService.Get(uint(idAsUInt))

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			var updateAppPayload application.ApplicationUpdate
			err = decodeJSONBody(rw, r, &updateAppPayload)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			app.ManifestPath = updateAppPayload.ManifestPath
			app.RepoID = updateAppPayload.RepoID
			app.Name = updateAppPayload.Name
			applicationService.Update(app)

			repo, err := repoService.Get(app.RepoID)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			repoDirRequest := reposerver.RepoDirRequest{RepoUrl: repo.Url}
			repoDirResponse, err := rp.GetRepoDir(context.Background(), &repoDirRequest)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			fullManifestPath := path.Join(repoDirResponse.Path, app.ManifestPath)
			message := reposerver.ManifestsRequest{Path: fullManifestPath}
			response, err := rp.GetManifests(context.Background(), &message)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			resp.App = app
			resp.Manifests = response
			respBytes, err := json.Marshal(resp)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(respBytes))
			return
		case "DELETE":
			app, err := applicationService.Get(uint(idAsUInt))

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			err = applicationService.Delete(app)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			http.Error(rw, "", http.StatusNoContent)
			io.WriteString(rw, "")
		default:
			JSONError(rw, errorResp{Message: "Something went wrong..."}, http.StatusInternalServerError)
		}
	}
}

func applicationsHandler(service *application.Service) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			apps := service.List()
			appsBytes, err := json.Marshal(apps)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(appsBytes))
		case "POST":
			var newAppPayload application.Application
			err := decodeJSONBody(rw, r, &newAppPayload)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			newApp := service.Create(newAppPayload)
			newAppBytes, err := json.Marshal(newApp)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(newAppBytes))
		default:
			JSONError(rw, errorResp{Message: "Something went wrong..."}, http.StatusInternalServerError)
		}
	}
}

func (s *Server) settingsHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		conn, err := grpc.Dial(":9000", grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Errorln(err)
		}

		defer conn.Close()

		rp := reposerver.NewRepoServiceClient(conn)

		switch r.Method {
		case "GET":
			message := reposerver.PathsRequest{}
			paths, err := rp.GetPaths(context.Background(), &message)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			resp := SettingsHttpResponse{
				RepoRoot:   paths.RepoRoot,
				SshRoot:    paths.SshRoot,
				PrivateKey: paths.PrivateKey,
			}
			respBytes, err := json.Marshal(resp)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(respBytes))
		default:
			JSONError(rw, errorResp{Message: "Something went wrong..."}, http.StatusInternalServerError)
		}
	}
}

type RetainedLogs struct {
	Logs []string `json:"logs"`
}

type FileData struct {
	DateCreated time.Time `json:"date_created"`
	Path        string    `json:"path"`
	Logs        []string  `json:"logs"`
}

type Pair struct {
	Key   string   `json:"file"`
	Value FileData `json:"file_data"`
}

type PairList []Pair

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value.DateCreated.Before(p[j].Value.DateCreated) }

func (s *Server) logsHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			vars := mux.Vars(r)
			jobId := vars["id"]
			libRegEx, err := regexp.Compile(`^.+.(log)$`)

			if err != nil {
				log.Errorln(err)
			}

			logFiles := make(map[string]FileData)

			err = filepath.Walk(path.Join(s.logsPath, jobId), func(path string, info os.FileInfo, err error) error {
				if err == nil {
					fileName := info.Name()
					if libRegEx.MatchString(fileName) {
						_, _, ctime, err := statTimes(path)

						if err != nil {
							return err
						}

						logLines, err := readLines(path)

						if err != nil {
							return err
						}

						logFiles[fileName] = FileData{
							DateCreated: ctime,
							Path:        path,
							Logs:        logLines,
						}
					}
				}

				return nil
			})

			if err != nil {
				log.Errorln(err)
			}

			p := make(PairList, len(logFiles))
			i := 0

			for k, v := range logFiles {
				p[i] = Pair{k, v}
				i++
			}

			sort.Sort(p)
			bytes, _ := json.Marshal(p)
			io.WriteString(rw, string(bytes))
		default:
			JSONError(rw, errorResp{Message: "Something went wrong..."}, http.StatusInternalServerError)
		}
	}
}

func (s *Server) applicationJobHandler(applicationService *application.Service, jobService *job.JobService) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		applicationID := vars["id"]
		idAsUInt, err := strconv.ParseUint(applicationID, 10, 32)

		if err != nil {
			log.Errorln(err)
		}

		switch r.Method {
		case "GET":
			vars := mux.Vars(r)
			applicationId := vars["id"]
			idAsUInt, err := strconv.ParseUint(applicationId, 10, 32)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			jobs, err := jobService.GetAllByAppId(uint(idAsUInt))

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			respBytes, err := json.Marshal(jobs)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(respBytes))
			return
		case "POST":
			var jobPayload client.JobConfig
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&jobPayload)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			randomString, err := GenerateRandomString(10)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			jobName := fmt.Sprintf("%s-%s", jobPayload.ObjectMeta.Name, randomString)
			jobConfig := client.JobConfig{ObjectMeta: jobPayload.ObjectMeta, Spec: jobPayload.Spec}
			newJob := jobService.Create(job.Job{Name: jobName, ApplicationID: uint(idAsUInt)})
			labels := make(map[string]string)

			labels["invoked"] = ""
			labels["job_id"] = strconv.FormatUint(uint64(newJob.ID), 10)

			jobConfig.Labels = labels
			resp, err := s.pods.Run(jobName, jobConfig)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			spec, _ := json.Marshal(resp.Spec)
			meta, _ := json.Marshal(jobPayload.ObjectMeta)
			newJob.Spec = spec
			newJob.Meta = meta
			jobService.Update(newJob)

			respBytes, err := json.Marshal(JobRunResponse{
				Job:    newJob,
				Config: jobConfig,
				Spec:   resp.Spec,
				Status: resp.Status,
			})

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(respBytes))
		default:
			JSONError(rw, errorResp{Message: "Something went wrong..."}, http.StatusInternalServerError)
		}
	}
}

func (s *Server) jobHandler(jobService *job.JobService) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			vars := mux.Vars(r)
			jobId := vars["id"]
			idAsUInt, err := strconv.ParseUint(jobId, 10, 32)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			job, err := jobService.Get(uint(idAsUInt))

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			respBytes, err := json.Marshal(job)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			io.WriteString(rw, string(respBytes))
		case "DELETE":
			vars := mux.Vars(r)
			jobId := vars["id"]
			idAsUInt, err := strconv.ParseUint(jobId, 10, 32)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			app, err := jobService.Get(uint(idAsUInt))

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			err = jobService.Delete(app)

			if err != nil {
				JSONError(rw, errorResp{Message: err.Error()}, http.StatusInternalServerError)
				return
			}

			http.Error(rw, "", http.StatusNoContent)
			io.WriteString(rw, "")
			io.WriteString(rw, string(""))
		default:
			JSONError(rw, errorResp{Message: "Something went wrong..."}, http.StatusInternalServerError)
		}
	}
}
