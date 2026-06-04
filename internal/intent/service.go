package intent

type Service struct {
	repo     Repository
	detector Detector
}

func NewService(repo Repository) Service {
	return Service{repo: repo, detector: NewDetector()}
}

func (s Service) Detect(projectID, content string) (Profile, error) {
	return s.repo.SaveProfile(s.detector.Detect(projectID, content))
}

func (s Service) Get(id string) (Profile, error) { return s.repo.GetProfile(id) }

func (s Service) Latest(projectID string) (Profile, error) { return s.repo.LatestProfile(projectID) }

func (s Service) Approve(id string) (Profile, error) { return s.repo.ApproveProfile(id) }
