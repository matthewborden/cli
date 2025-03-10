package factory

import (
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/buildkite/cli/v3/internal/config"
	"github.com/buildkite/cli/v3/internal/version"
	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/go-git/go-git/v5"
	"github.com/sashabaranov/go-openai"
)

var userAgent string

type Factory struct {
	Config        *config.Config
	GitRepository *git.Repository
	GraphQLClient graphql.Client
	OpenAIClient  *openai.Client
	RestAPIClient *buildkite.Client
	Version       string
}

type gqlHTTPClient struct {
	client *http.Client
	token  string
}

func init() {
	userAgent = fmt.Sprintf("%s buildkite-cli/%s", buildkite.DefaultUserAgent, version.Version)
}

func (a *gqlHTTPClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.token))
	req.Header.Set("User-Agent", userAgent)
	return a.client.Do(req)
}

func New(version string) (*Factory, error) {
	repo, _ := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	conf := config.New(nil, repo)
	buildkiteClient, err := buildkite.NewOpts(
		buildkite.WithTokenAuth(conf.APIToken()),
		buildkite.WithUserAgent(userAgent),
	)
	if err != nil {
		return nil, fmt.Errorf("creating buildkite client: %w", err)
	}

	graphqlHTTPClient := &gqlHTTPClient{client: http.DefaultClient, token: conf.APIToken()}

	return &Factory{
		Config:        conf,
		GitRepository: repo,
		GraphQLClient: graphql.NewClient(config.DefaultGraphQLEndpoint, graphqlHTTPClient),
		OpenAIClient:  openai.NewClient(conf.GetOpenAIToken()),
		RestAPIClient: buildkiteClient,
		Version:       version,
	}, nil
}
