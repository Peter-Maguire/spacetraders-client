package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"io"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type IncomingResponse struct {
	Response *http.Response
	Data     []byte
	Error    error
}

type OutgoingRequest struct {
	Req            *http.Request
	OriginalPath   string
	ReturnChannels []chan IncomingResponse
	Mutex          sync.Mutex
	Priority       int
	Context        context.Context
}

var RequestBuffer = make([]*OutgoingRequest, 0)

var RBufferLock sync.Mutex

var Waiting = 0

var IsRunningRequests = false

var (
	httpResponses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "st_http_responses",
		Help: "Number of HTTP responses by code",
	}, []string{"path", "method", "code"})

	httpRequestTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "st_http_response_time",
		Help: "HTTP response time",
	})
)

func makeRequest[T any](ctx context.Context, method string, path string, body any) (*HttpResponse[T], *HttpError) {
	// We can't set this to bytes.Buffer type because net/http assumes data of that type will not be nil
	var buf io.Reader = nil
	if body != nil {
		data, err := json.Marshal(body)
		//fmt.Println(string(data))
		if err != nil {
			return nil, InternalError(err)
		}
		buf = bytes.NewBuffer(data)
	}
	if ctx.Value("state") != nil {
		ctx.Value("state").(State).SetWaitingForHttp(true)
	}
	req, err := http.NewRequest(method, fmt.Sprintf("https://api.spacetraders.io/v2/%s", path), buf)
	if err != nil {
		return nil, InternalError(err)
	}
	if body != nil {
		req.Header.Set("content-type", "application/json")
	}

	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", ctx.Value("token")))
	returnChan := make(chan IncomingResponse)

	usingExistingRequest := false

	if method == "GET" {
		RBufferLock.Lock()
		for _, bufferedRequest := range RequestBuffer {
			if bufferedRequest.Req.Method == "GET" && bufferedRequest.OriginalPath == path && bufferedRequest.Req.Header.Get("Authorization") == req.Header.Get("Authorization") {
				bufferedRequest.Mutex.Lock()
				bufferedRequest.ReturnChannels = append(bufferedRequest.ReturnChannels, returnChan)
				bufferedRequest.Mutex.Unlock()
				fmt.Println("Found request to piggyback on")
				bufferedRequest.Priority += 10
				usingExistingRequest = true
				break
			}
		}
		RBufferLock.Unlock()
	}
	Waiting++
	if !usingExistingRequest {
		RBufferLock.Lock()
		RequestBuffer = append(RequestBuffer, &OutgoingRequest{
			Req:            req,
			ReturnChannels: []chan IncomingResponse{returnChan},
			Priority:       getRequestPriority(path),
			OriginalPath:   path,
			Context:        ctx,
		})
		RBufferLock.Unlock()
		if !IsRunningRequests {
			requestLoop()
		}
	}
	resp := <-returnChan
	Waiting--
	if resp.Error != nil {
		return nil, InternalError(resp.Error)
	}
	output := &HttpResponse[T]{}
	fmt.Printf("%s: %s\n", req.Method, req.URL.Path)
	fmt.Println(string(resp.Data))
	err = json.Unmarshal(resp.Data, output)
	if output.Error != nil {
		return output, output.Error
	}
	if err != nil {
		return output, InternalError(err)
	}
	return output, nil
}

func PaginatedRequest[T any](ctx context.Context, path string, startPage int, maxPages int) (*[]T, *HttpError) {
	var currentPage = startPage
	output := make([]T, 0)
	for {
		fixedPath := path
		hasQueryString := strings.Contains(path, "?")
		if hasQueryString {
			fixedPath += "&"
		} else {
			fixedPath += "?"
		}
		fixedPath += fmt.Sprintf("limit=20&page=%d", currentPage)

		resp, err := makeRequest[[]T](ctx, "GET", fixedPath, nil)
		if err != nil {
			return &output, err
		}
		output = append(output, *resp.Data...)
		if maxPages > 0 && resp.PaginatedMeta.Page >= maxPages {
			// Reached max page count
			return &output, nil
		}

		if (resp.PaginatedMeta.Page)*resp.PaginatedMeta.Limit >= resp.PaginatedMeta.Total || len(*resp.Data) == 0 {
			// Reached last page
			return &output, nil
		}
		currentPage++
	}

}

func Request[T any](ctx context.Context, method string, path string, body any) (*T, *HttpError) {
	resp, err := makeRequest[T](ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Data, nil
}

func doRequests() (bool, time.Duration) {
	if len(RequestBuffer) == 0 {
		return false, 0
	}
	requestStart := time.Now()
	RBufferLock.Lock()
	or := RequestBuffer[0]
	RequestBuffer = RequestBuffer[1:]
	RBufferLock.Unlock()
	res, err := http.DefaultClient.Do(or.Req)
	requestStop := time.Now()
	var data []byte
	if or.Context.Value("state") != nil {
		or.Context.Value("state").(State).SetWaitingForHttp(false)
	}
	if err == nil {
		data, err = io.ReadAll(res.Body)
		httpResponses.WithLabelValues(or.OriginalPath, or.Req.Method, strconv.Itoa(res.StatusCode)).Inc()
	}

	for _, ch := range or.ReturnChannels {
		ch <- IncomingResponse{
			Response: res,
			Error:    err,
			Data:     data,
		}
	}

	requestTime := requestStop.Sub(requestStart)
	httpRequestTime.Observe(float64(requestTime))
	return true, requestTime
}

func requestLoop() {
	IsRunningRequests = true
	go func() {
		for {
			sort.SliceStable(RequestBuffer, func(i, j int) bool {
				return RequestBuffer[i].Priority > RequestBuffer[j].Priority
			})
			requests, timing := doRequests()
			if !requests {
				IsRunningRequests = false
				break
			}
			time.Sleep((500 * time.Millisecond) - timing)
		}
	}()
}

var priorities = map[string]int{
	// Navigation is top priority as it takes the longest
	"jump":     16,
	"warp":     16,
	"navigate": 15,
	"nav":      15,
	"refuel":   14,
	"orbit":    14,
	"survey":   11,
	// Mining should happen before other things
	"extract":  10,
	"jettison": 10,
	"refine":   10,
	"transfer": 10,
	// Fulfilling contracts or construction sites
	"supply":  9,
	"fulfill": 9,
	"deliver": 9,
	// Selling has to have priority over transfers for haulers
	"sell":  2,
	"cargo": 2,
	// Scan has priority over most things
	"scan": 2,
}

func getRequestPriority(requestPath string) int {
	priority, ok := priorities[path.Base(requestPath)]
	if ok {
		return priority
	}
	return 1
}

func Init() {

}
