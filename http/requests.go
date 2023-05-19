package http

import (
    "bytes"
    "encoding/json"
    "fmt"
    "github.com/jellydator/ttlcache/v3"
    "io"
    "net/http"
    "os"
    "sort"
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
}

var cache = ttlcache.New[string, http.Response](
    ttlcache.WithTTL[string, http.Response](30 * time.Minute),
)

var RequestBuffer = make([]*OutgoingRequest, 0)

var RBufferLock sync.Mutex

var Waiting = 0

var IsRunningRequests = false

var token = fmt.Sprintf("Bearer %s", os.Getenv("TOKEN"))

func makeRequest[T any](method string, path string, body any) (*HttpResponse[T], *HttpError) {
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
    req, err := http.NewRequest(method, fmt.Sprintf("https://api.spacetraders.io/v2/%s", path), buf)
    if err != nil {
        return nil, InternalError(err)
    }
    if body != nil {
        req.Header.Set("content-type", "application/json")
    }
    req.Header.Add("authorization", token)
    returnChan := make(chan IncomingResponse)

    usingExistingRequest := false

    if method == "GET" {
        RBufferLock.Lock()
        for _, bufferedRequest := range RequestBuffer {
            if bufferedRequest.Req.Method == "GET" && bufferedRequest.OriginalPath == path {
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
        })
        RBufferLock.Unlock()
        if !IsRunningRequests {
            requestLoop()
        }
    }
    resp := <-returnChan
    Waiting--
    if err != nil {
        return nil, InternalError(err)
    }
    if resp.Error != nil {
        return nil, InternalError(resp.Error)
    }
    output := &HttpResponse[T]{}
    err = json.Unmarshal(resp.Data, output)
    if output.Error != nil {
        return output, output.Error
    }
    if err != nil {
        return output, InternalError(err)
    }
    return output, nil
}

func PaginatedRequest[T any](path string, startPage int, maxPages int) (*[]T, *HttpError) {
    var currentPage = startPage
    output := make([]T, 0)
    for {
        resp, err := makeRequest[[]T]("GET", fmt.Sprintf("%s?limit=20&page=%d", path, currentPage), nil)
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

func Request[T any](method string, path string, body any) (*T, *HttpError) {
    resp, err := makeRequest[T](method, path, body)
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

    var data []byte

    if err == nil {
        data, err = io.ReadAll(res.Body)
    }

    for _, ch := range or.ReturnChannels {
        ch <- IncomingResponse{
            Response: res,
            Error:    err,
            Data:     data,
        }
    }

    requestStop := time.Now()
    requestTime := requestStop.Sub(requestStart)
    return true, requestTime
}

func requestLoop() {
    IsRunningRequests = true
    go func() {
        for {
            sort.Slice(RequestBuffer, func(i, j int) bool {
                return RequestBuffer[i].Priority > RequestBuffer[j].Priority
            })
            requests, timing := doRequests()
            if !requests {
                IsRunningRequests = false
                break
            }
            time.Sleep(time.Second - timing)
        }
    }()
}

func getRequestPriority(path string) int {
    // Navigation is top priority as it takes the longest
    if strings.HasSuffix(path, "/jump") || strings.HasSuffix(path, "/warp") {
        return 16
    }
    if strings.HasSuffix(path, "/navigate") {
        return 15
    }
    // Survey should happen before mining
    if strings.HasSuffix(path, "/survey") {
        return 11
    }
    // Mining should happen before other things
    if strings.HasSuffix(path, "/extract") || strings.HasSuffix(path, "/jettison") {
        return 10
    }
    // Selling has to have priority over transfers for haulers
    if strings.HasSuffix(path, "/sell") || strings.HasSuffix(path, "/cargo") {
        return 2
    }
    return 1
}

func Init() {
    go cache.Start()
}
