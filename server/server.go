package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/infor-design/selfservice/pkg/client"
	"github.com/infor-design/selfservice/pkg/db"
	"github.com/infor-design/selfservice/pkg/health"
	"github.com/infor-design/selfservice/pkg/job"
	"github.com/infor-design/selfservice/pkg/repo"
	"github.com/infor-design/selfservice/pkg/utils"
	"github.com/infor-design/selfservice/reposerver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/infor-design/selfservice/pkg/application"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type ServerConfig struct{}

type Server struct {
	ServerConfig
	log             *log.Entry
	logsPath        string
	refreshInterval int
	db              *db.Connection
	clientset       *client.Clientset
	pods            *client.Client
	repoService     *repo.Service
	router          *mux.Router
	stopCh          chan struct{}
}

func NewServer(config ServerConfig) *Server {
	dbConfig := db.NewConfig()
	newDb := db.NewDb(dbConfig)

	return &Server{
		ServerConfig:    config,
		db:              newDb,
		log:             log.NewEntry(log.StandardLogger()),
		logsPath:        utils.GetEnv("LOGS_PATH", ""),
		refreshInterval: 15,
		clientset:       client.NewClientset(),
		pods:            client.NewClient(),
		repoService:     repo.NewService(newDb),
		router:          mux.NewRouter().StrictSlash(true),
	}
}

func (s *Server) Init() {
	s.db.InitialMigration()
}

func (s *Server) Run() {
	httpState := health.NewState()
	jobService := job.NewService(s.db)
	applicationService := application.NewService(s.db)
	informer := client.NewInformer(s.clientset, jobService)

	s.router.HandleFunc("/repos", reposHandler(s.repoService))
	s.router.HandleFunc("/repos/{id:[0-9]+}", repoHandler(s.repoService))
	s.router.HandleFunc("/repos/{id:[0-9]+}/{action:[a-z]+}", repoHandler(s.repoService))

	s.router.HandleFunc("/applications", applicationsHandler(applicationService))
	s.router.HandleFunc("/applications/{id:[0-9]+}", applicationHandler(applicationService, s.repoService))
	s.router.HandleFunc("/applications/{id:[0-9]+}/jobs", s.applicationJobHandler(applicationService, jobService))

	s.router.HandleFunc("/jobs/{id:[0-9]+}", s.jobHandler(jobService))
	s.router.HandleFunc("/jobs/{id:[0-9]+}/logs", s.logsHandler())

	s.router.HandleFunc("/settings", s.settingsHandler())
	s.router.HandleFunc("/health", httpState.Health)

	s.router.Use(contentTypeApplicationJsonMiddleware)
	s.router.Use(corsMiddleware)
	http.Handle("/", s.router)

	conn, err := grpc.Dial(":9000", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("%v", err)
	}

	defer conn.Close()

	rp := reposerver.NewRepoServiceClient(conn)

	go func() {
		ticker := time.NewTicker(time.Minute * time.Duration(s.refreshInterval))
		defer ticker.Stop()

		for range ticker.C {
			repos := s.repoService.List()

			for _, repo := range repos {
				message := reposerver.SyncRequest{Repo: repo.Url, RepoId: strconv.FormatInt(int64(repo.ID), 10)}
				resp, err := rp.Sync(context.Background(), &message)

				if err != nil {
					log.Errorln(err)
					continue
				}

				updateRepo, err := s.repoService.Get(uint(repo.ID))

				if err != nil {
					log.Errorln(err)
					continue
				}

				updateRepo.Commit = resp.Commit
				updateRepo.Hash = resp.Hash
				s.repoService.Update(updateRepo)
			}
		}
	}()

	go informer.StartInformer()
	go func() {
		log.Infof("Starting server...")
		s.checkServeErr("http", http.ListenAndServe(":8080", nil))
	}()

	s.stopCh = make(chan struct{})
	<-s.stopCh
}

func (a *Server) checkServeErr(name string, err error) {
	if err != nil {
		if a.stopCh == nil {
			log.Infof("graceful shutdown %s: %v", name, err)
		} else {
			log.Fatalf("%s: %v", name, err)
		}
	} else {
		log.Infof("graceful shutdown %s", name)
	}
}

func (s *Server) Shutdown() {
	log.Info("Shut down requested")
	stopCh := s.stopCh
	s.stopCh = nil
	s.db.Close()
	if stopCh != nil {
		close(stopCh)
	}
}
