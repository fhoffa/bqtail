package dispatch

import (
	"bqtail/base"
	"bqtail/dispatch/contract"
	"bqtail/service/bq"
	"bqtail/service/storage"
	"bqtail/task"
	"bytes"
	"context"
	"encoding/json"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/url"
	"github.com/viant/toolbox"
	"google.golang.org/api/bigquery/v2"
	"path"
	"strings"
)

//Service represents event service
type Service interface {
	Dispatch(ctx context.Context, request *contract.Request) *contract.Response
}

type service struct {
	task.Registry
	config  *Config
	bq      bq.Service
	storage afs.Service
}

func (s *service) Init(ctx context.Context) error {
	err := s.config.Init(ctx)
	if err != nil {
		return err
	}
	bqService, err := bigquery.NewService(ctx)
	if err != nil {
		return err
	}
	s.bq = bq.New(bqService, s.Registry, s.config.ProjectID, s.storage)
	bq.InitRegistry(s.Registry, s.bq)
	storage.InitRegistry(s.Registry, storage.New(s.storage))
	return err
}

func (s *service) Dispatch(ctx context.Context, request *contract.Request) *contract.Response {
	response := contract.NewResponse(request.EventID)
	defer response.SetTimeTaken(response.Started)
	err := s.dispatch(ctx, request, response)
	if err != nil {
		response.SetIfError(err)
	}
	return response
}

func (s *service) initRequest(ctx context.Context, request *contract.Request) error {
	job, err := s.bq.GetJob(ctx, request.ProjectID, request.JobID)
	if err != nil {
		return err
	}
	contractJob := contract.Job(*job)
	request.Job = &contractJob
	return nil
}

//move moves shedule file to output folder
func (s *service) move(ctx context.Context, baseURL string, job *Job) error {
	matchedURL := job.Response.MatchedURL
	if matchedURL == "" {
		return nil
	}
	parent, sourceName := url.Split(matchedURL, file.Scheme)
	parentElements := strings.Split(parent, "/")
	if len(parentElements) > 2 {
		sourceName = path.Join(strings.Join(parentElements[len(parentElements)-2:], "/"), sourceName)
	}
	name := path.Join(job.Completed().Format(dateLayout), sourceName)
	URL := url.Join(baseURL, name)
	return s.storage.Move(ctx, matchedURL, URL)
}

func (s *service) onDone(ctx context.Context, job *Job) error {
	baseURL := s.config.OutputURL(job.Response.Status == base.StatusError)
	if err := s.move(ctx, baseURL, job); err != nil {
		return err
	}
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	jobFilename := path.Join(base.DecodeID(job.JobReference.JobId))
	name := path.Join(job.Completed().Format(dateLayout), jobFilename+base.JobExt+"-done")
	URL := url.Join(baseURL, name)
	return s.storage.Upload(ctx, URL, file.DefaultFileOsMode, bytes.NewReader(data))
}

func (s *service) getActions(ctx context.Context, request *contract.Request, response *contract.Response) (*task.Actions, error) {
	jobID := base.DecodeID(request.JobID)
	if strings.HasSuffix(jobID, base.BqDispatchJob) {
		URL := url.Join(s.config.DispatchURL, jobID+base.JobExt)
		response.MatchedURL = URL
		response.Matched = true
		reader, err := s.storage.DownloadWithURL(ctx, URL)
		if err != nil {
			return nil, nil
		}
		defer func() { _ = reader.Close() }()
		actions := &task.Actions{}
		return actions, json.NewDecoder(reader).Decode(actions)
	}

	return nil, nil
}

func (s *service) dispatch(ctx context.Context, request *contract.Request, response *contract.Response) (err error) {
	err = s.initRequest(ctx, request)
	if err != nil {
		return err
	}
	job := NewJob(request.Job, response)
	defer func() {
		if response.Matched || err != nil {
			response.SetIfError(err)
			if e := s.onDone(ctx, job); e != nil && err == nil {
				e = err
			}
		}
	}()
	response.JobRef = request.Job.JobReference
	if err := request.Job.Error(); err != nil {
		response.JobError = err.Error()
	}
	route := s.config.Routes.Match(request.Job)
	if route != nil {
		expandable := &base.Expandable{}
		if route.When.Dest != "" {
			expandable.Source = request.Job.Dest()
		} else if route.When.Source != "" {
			expandable.Source = request.Job.Source()
		}
		response.Matched = true
		job.Actions = route.Actions.Expand(expandable)
		toolbox.Dump(job.Actions)
		return s.run(ctx, job)
	}
	var actions *task.Actions
	job.Actions, err = s.getActions(ctx, request, response)
	if err != nil || job.Actions == nil || job.Actions.IsEmpty() {
		return err
	}
	job.Actions = actions
	response.Matched = true
	err = s.run(ctx, job)
	return err
}

func (s *service) run(ctx context.Context, job *Job) error {
	toRun := job.ToRun()
	var err error
	for i := range toRun {
		if err = task.Run(ctx, s.Registry, toRun[i]); err != nil {
			return err
		}
	}
	return err
}

//New creates a dispatch service
func New(ctx context.Context, config *Config) (Service, error) {
	srv := &service{
		config:   config,
		storage:  afs.New(),
		Registry: task.NewRegistry(),
	}
	return srv, srv.Init(ctx)
}