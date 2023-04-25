package client

import (
	"bytes"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"golang.org/x/net/context"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	doer    doer
	method  string
	baseUrl string
}

func NewClient(d doer, method, baseUrl string) *Client {
	return &Client{doer: d, method: method, baseUrl: baseUrl}
}

func (c *Client) SetUrl(url string) {
	c.baseUrl = url
}

func (c *Client) SetMethod(method string) {
	c.method = method
}

func (c *Client) GetClientUrl() string {
	return c.baseUrl
}

// QueryRange 范围查询
func (c *Client) QueryRange(ctx context.Context, q *models.Query, headers http.Header) (*http.Response, error) {
	tr := q.TimeRange()

	// 构建prometheus查询url
	u, err := c.createUrl("api/v1/query_range", map[string]string{
		"query": q.Expr,
		"start": formatTime(tr.Start),
		"end":   formatTime(tr.End),
		"step":  strconv.FormatFloat(tr.Step.Seconds(), 'f', -1, 64),
	})
	if err != nil {
		return nil, err
	}

	log.DefaultLogger.Info("Header is: ", headers)
	req, err := createRequest(ctx, c.method, u, nil, headers)

	if err != nil {
		return nil, err
	}
	return c.doer.Do(req)
}

// QueryInstant 实时查询
func (c *Client) QueryInstant(ctx context.Context, q *models.Query, headers http.Header) (*http.Response, error) {
	qs := map[string]string{"query": q.Expr}
	tr := q.TimeRange()
	if !tr.End.IsZero() {
		qs["time"] = formatTime(tr.End)
	}

	u, err := c.createUrl("api/v1/query", qs)
	if err != nil {
		return nil, err
	}
	req, err := createRequest(ctx, c.method, u, nil, headers)
	if err != nil {
		return nil, err
	}
	return c.doer.Do(req)
}

// QueryExemplars 查询调用链
func (c *Client) QueryExemplars(ctx context.Context, q *models.Query, headers http.Header) (*http.Response, error) {
	tr := q.TimeRange()
	u, err := c.createUrl("api/v1/query_exemplars", map[string]string{
		"query": q.Expr,
		"start": formatTime(tr.Start),
		"end":   formatTime(tr.End),
	})
	if err != nil {
		return nil, err
	}

	req, err := createRequest(ctx, c.method, u, nil, headers)
	if err != nil {
		return nil, err
	}
	return c.doer.Do(req)
}

func (c *Client) QueryMetrics(ctx context.Context, start int64, end int64, headers http.Header) (*http.Response,
	error) {
	u, err := c.createUrl("api/v1/label/__name__/values", map[string]string{
		"query": "__name__/values",
		//"start": formatTime(time.Unix(start, 0)),
		//"end":   formatTime(time.Unix(end, 0)),
	})
	if err != nil {
		return nil, err
	}

	req, err := createRequest(ctx, http.MethodGet, u, nil, headers)
	if err != nil {
		return nil, err
	}
	return c.doer.Do(req)
}

func (c *Client) QueryLabelNames(ctx context.Context, start int64, end int64, headers http.Header) (*http.Response,
	error) {
	u, err := c.createUrl("api/v1/labels", map[string]string{
		//"start": formatTime(time.Unix(start, 0)),
		//"end":   formatTime(time.Unix(end, 0)),
	})
	if err != nil {
		return nil, err
	}

	req, err := createRequest(ctx, http.MethodGet, u, nil, headers)
	if err != nil {
		return nil, err
	}
	return c.doer.Do(req)
}

func (c *Client) QuerySeries(ctx context.Context, start int64, end int64, match string,
	headers http.Header, needTime bool) (*http.Response, error) {
	var (
		u   *url.URL
		err error
	)
	if needTime {
		u, err = c.createUrl("api/v1/series", map[string]string{
			"start":   formatTime(time.Unix(start, 0)),
			"end":     formatTime(time.Unix(end, 0)),
			"match[]": match,
		})
	} else {
		u, err = c.createUrl("api/v1/series", map[string]string{
			"match[]": match,
		})
	}
	if err != nil {
		return nil, err
	}

	req, err := createRequest(ctx, http.MethodGet, u, nil, headers)
	if err != nil {
		return nil, err
	}
	return c.doer.Do(req)
}

func (c *Client) CallAlgorithm(ctx context.Context, body []byte, headers http.Header) (*http.Response, error) {
	log.DefaultLogger.Info("Http header is: ", headers)
	u, err := c.createUrl("/", map[string]string{})
	if err != nil {
		return nil, err
	}
	req, err := createRequest(ctx, c.method, u, body, headers)
	if err != nil {
		return nil, err
	}
	return c.doer.Do(req)
}

func (c *Client) GetAlgorithmList(ctx context.Context, headers http.Header) (*http.Response, error) {
	u, err := c.createUrl("/", map[string]string{})
	if err != nil {
		return nil, err
	}
	req, err := createRequest(ctx, "GET", u, nil, headers)
	if err != nil {
		return nil, err
	}
	return c.doer.Do(req)
}

func (c *Client) CheckHealthy(ctx context.Context, headers http.Header) (*http.Response, error) {
	u, err := c.createUrl("/", map[string]string{})
	if err != nil {
		return nil, err
	}
	req, err := createRequest(ctx, c.method, u, nil, headers)
	if err != nil {
		return nil, err
	}
	return c.doer.Do(req)
}

// createUrl 构建prometheus查询url
func (c *Client) createUrl(endpoint string, qs map[string]string) (*url.URL, error) {
	finalUrl, err := url.ParseRequestURI(c.baseUrl)
	if err != nil {
		return nil, err
	}

	finalUrl.Path = path.Join(finalUrl.Path, endpoint)
	urlQuery := finalUrl.Query()

	for key, val := range qs {
		urlQuery.Set(key, val)
	}
	finalUrl.RawQuery = urlQuery.Encode()
	log.DefaultLogger.Info("Final url is:", finalUrl)
	log.DefaultLogger.Info("Final url is:", finalUrl.String())
	return finalUrl, nil
}

func formatTime(t time.Time) string {
	return strconv.FormatFloat(float64(t.Unix())+float64(t.Nanosecond())/1e9, 'f', -1, 64)
}

func createRequest(ctx context.Context, method string, u *url.URL, body []byte, header http.Header) (*http.Request,
	error) {
	bodyReader := bytes.NewReader(body)
	request, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		log.DefaultLogger.Error("Request error, error is: ", err)
		return nil, err
	}

	if header != nil {
		request.Header = header
	}

	if strings.ToUpper(method) == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
		request.Header["Idempotency-Key"] = nil
	}
	return request, nil
}
