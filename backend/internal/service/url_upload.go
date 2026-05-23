package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"omepic/backend/internal/util"
)

const (
	defaultURLUploadTimeout     = 30 * time.Second
	defaultURLUploadRedirectMax = 5
)

type URLUploadInput struct {
	URL        string
	Token      string
	IPAddress  string
	BaseURL    string
	StorageKey string
}

type RemoteImageFetcher struct {
	Client       *http.Client
	MaxRedirects int
}

type remoteImageDownload struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
	Filename      string
}

func (s *ImageService) SetRemoteImageFetcher(fetcher *RemoteImageFetcher) {
	s.remoteImageFetcher = fetcher
}

func (s *ImageService) UploadRemoteURL(ctx context.Context, input URLUploadInput) (UploadOutput, error) {
	fetcher := s.remoteImageFetcher
	if fetcher == nil {
		fetcher = NewRemoteImageFetcher()
	}
	downstream, err := fetcher.Fetch(ctx, input.URL, s.MaxUploadSizeBytes())
	if err != nil {
		return UploadOutput{}, err
	}
	defer downstream.Body.Close()

	return s.Upload(ctx, UploadInput{
		Token:            input.Token,
		OriginalFilename: downstream.Filename,
		MIMEType:         downstream.ContentType,
		IPAddress:        input.IPAddress,
		Source:           downstream.Body,
		DeclaredSize:     downstream.ContentLength,
		BaseURL:          input.BaseURL,
		StorageKey:       input.StorageKey,
	})
}

func NewRemoteImageFetcher() *RemoteImageFetcher {
	return &RemoteImageFetcher{Client: safeURLUploadHTTPClient(), MaxRedirects: defaultURLUploadRedirectMax}
}

func (f *RemoteImageFetcher) Fetch(ctx context.Context, rawURL string, maxBytes int64) (remoteImageDownload, error) {
	parsed, err := util.ValidateRemoteImageURL(rawURL)
	if err != nil {
		return remoteImageDownload{}, WithUserMessage(ErrInvalidInput, "remote image url is not allowed")
	}
	if err := validateURLHostLiteral(parsed.String()); err != nil {
		return remoteImageDownload{}, err
	}
	client := f.Client
	if client == nil {
		client = safeURLUploadHTTPClient()
	}
	maxRedirects := f.MaxRedirects
	if maxRedirects <= 0 {
		maxRedirects = defaultURLUploadRedirectMax
	}
	redirects := 0
	clientCopy := *client
	clientCopy.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		redirects++
		if redirects > maxRedirects {
			return WithUserMessage(ErrInvalidInput, "too many redirects")
		}
		if err := validateURLHostLiteral(req.URL.String()); err != nil {
			return err
		}
		return nil
	}

	requestCtx, cancel := context.WithTimeout(ctx, defaultURLUploadTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return remoteImageDownload{}, WithUserMessage(ErrInvalidInput, "remote image url is invalid")
	}
	res, err := clientCopy.Do(req)
	if err != nil {
		if errorsIsUserInput(err) {
			return remoteImageDownload{}, err
		}
		if requestCtx.Err() != nil {
			return remoteImageDownload{}, fmt.Errorf("%w: remote image download timed out", ErrDependencyUnavailable)
		}
		return remoteImageDownload{}, WithUserMessage(ErrInvalidInput, "remote image url is not allowed")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		_ = res.Body.Close()
		return remoteImageDownload{}, WithUserMessage(ErrInvalidInput, "remote image download failed")
	}
	if maxBytes > 0 && res.ContentLength > maxBytes {
		_ = res.Body.Close()
		return remoteImageDownload{}, WithUserMessage(ErrInvalidInput, fmt.Sprintf("file size must be between 1 byte and %d MB", maxBytes/(1024*1024)))
	}
	contentType := strings.TrimSpace(res.Header.Get("Content-Type"))
	if mediaType, _, err := mime.ParseMediaType(contentType); err == nil {
		contentType = mediaType
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	body := io.ReadCloser(res.Body)
	if maxBytes > 0 {
		body = &limitedReadCloser{Reader: io.LimitReader(res.Body, maxBytes+1), Closer: res.Body}
	}
	return remoteImageDownload{Body: body, ContentType: contentType, ContentLength: res.ContentLength, Filename: filenameFromRemoteURL(res.Request.URL)}, nil
}

func safeURLUploadHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network string, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := util.ResolveAndValidateHost(ctx, net.DefaultResolver, host)
		if err != nil {
			return nil, err
		}
		var lastErr error
		for _, ip := range ips {
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
		return nil, lastErr
	}
	return &http.Client{Timeout: defaultURLUploadTimeout, Transport: transport}
}

type limitedReadCloser struct {
	io.Reader
	io.Closer
}

func validateURLHostLiteral(rawURL string) error {
	parsed, err := util.ValidateRemoteImageURL(rawURL)
	if err != nil {
		return WithUserMessage(ErrInvalidInput, "remote image url is not allowed")
	}
	if literal := net.ParseIP(parsed.Hostname()); literal != nil {
		if err := util.ValidateResolvedIP(literal); err != nil {
			return WithUserMessage(ErrInvalidInput, "remote image url is not allowed")
		}
	}
	return nil
}

func filenameFromRemoteURL(u *url.URL) string {
	name := strings.TrimSpace(path.Base(u.Path))
	if name == "" || name == "." || name == "/" {
		return "remote-image"
	}
	return name
}

func errorsIsUserInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}
