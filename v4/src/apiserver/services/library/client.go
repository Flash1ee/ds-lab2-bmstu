package library

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"

	"ds-lab2-bmstu/apiserver/core/ports/library"
	v1 "ds-lab2-bmstu/library/api/http/v1"
	"ds-lab2-bmstu/pkg/readiness"
	"ds-lab2-bmstu/pkg/readiness/httpprober"
)

const probeKey = "http-library-client"

var ErrInvalidStatusCode = errors.New("invalid status code")

type Client struct {
	lg *slog.Logger

	conn *resty.Client
}

func New(lg *slog.Logger, cfg library.Config, probe *readiness.Probe) (*Client, error) {
	client := resty.New().
		SetTransport(&http.Transport{
			MaxIdleConns:       10,               //nolint: gomnd
			IdleConnTimeout:    30 * time.Second, //nolint: gomnd
			DisableCompression: true,
		}).
		SetBaseURL(fmt.Sprintf("http://%s", net.JoinHostPort(cfg.Host, cfg.Port))).EnableTrace()

	c := Client{
		lg:   lg,
		conn: client,
	}

	go httpprober.New(lg, client).Ping(probeKey, probe)

	return &c, nil
}

func (c *Client) GetLibraries(
	_ context.Context, city string, page uint64, size uint64,
) (library.Infos, error) {
	q := map[string]string{
		"city": city,
		"page": strconv.FormatUint(page, 10),
	}

	if size == 0 {
		size = math.MaxUint64
	}

	q["size"] = strconv.FormatUint(size, 10)

	resp, err := c.conn.R().
		SetQueryParams(q).
		SetResult(&v1.LibrariesResponse{}).
		Get("/api/v1/libraries")
	if err != nil {
		return library.Infos{}, fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return library.Infos{}, fmt.Errorf("%d: %w", resp.StatusCode(), ErrInvalidStatusCode)
	}

	data, _ := resp.Result().(*v1.LibrariesResponse)

	libraries := library.Infos{Total: data.Total}
	for _, book := range data.Items {
		libraries.Items = append(libraries.Items, library.Info(book))
	}

	return libraries, nil
}

// nolint: dupl
func (c *Client) GetLibrariesByIDs(
	_ context.Context, ids []string,
) (library.Infos, error) {
	id, err := json.Marshal(ids)
	if err != nil {
		return library.Infos{}, fmt.Errorf("failed to marshal data: %w", err)
	}

	resp, err := c.conn.R().
		SetQueryParam("ids", string(id)).
		SetResult(&v1.LibrariesResponse{}).
		Get("/api/v1/libraries")
	if err != nil {
		return library.Infos{}, fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return library.Infos{}, fmt.Errorf("%d: %w", resp.StatusCode(), ErrInvalidStatusCode)
	}

	data, _ := resp.Result().(*v1.LibrariesResponse)

	libraries := library.Infos{Total: data.Total}
	for _, book := range data.Items {
		libraries.Items = append(libraries.Items, library.Info(book))
	}

	return libraries, nil
}

func (c *Client) GetBooks(
	_ context.Context, libraryID string, showAll bool, page uint64, size uint64,
) (library.Books, error) {
	if size == 0 {
		size = math.MaxUint64
	}

	q := map[string]string{
		"size": strconv.FormatUint(size, 10),
		"page": strconv.FormatUint(page, 10),
	}

	if showAll {
		q["show_all"] = "1"
	}

	resp, err := c.conn.R().
		SetQueryParams(q).
		SetPathParam("library_id", libraryID).
		SetResult(&v1.BooksResponse{}).
		Get("/api/v1/libraries/{library_id}/books")
	if err != nil {
		return library.Books{}, fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return library.Books{}, fmt.Errorf("%d: %w", resp.StatusCode(), ErrInvalidStatusCode)
	}

	data, _ := resp.Result().(*v1.BooksResponse)

	books := library.Books{Total: data.Total}
	for _, book := range data.Items {
		books.Items = append(books.Items, library.Book(book))
	}

	return books, nil
}

// nolint: dupl
func (c *Client) GetBooksByIDs(
	_ context.Context, ids []string,
) (library.Books, error) {
	id, err := json.Marshal(ids)
	if err != nil {
		return library.Books{}, fmt.Errorf("failed to marshal data: %w", err)
	}

	resp, err := c.conn.R().
		SetQueryParam("ids", string(id)).
		SetResult(&v1.BooksResponse{}).
		Get("/api/v1/books")
	if err != nil {
		return library.Books{}, fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return library.Books{}, fmt.Errorf("%d: %w", resp.StatusCode(), ErrInvalidStatusCode)
	}

	data, _ := resp.Result().(*v1.BooksResponse)

	books := library.Books{Total: data.Total}
	for _, book := range data.Items {
		books.Items = append(books.Items, library.Book(book))
	}

	return books, nil
}

func (c *Client) ObtainBook(
	_ context.Context, libraryID string, bookID string,
) (library.ReservedBook, error) {
	body, err := json.Marshal(v1.TakeBookRequest{
		BookID:    bookID,
		LibraryID: libraryID,
	})
	if err != nil {
		return library.ReservedBook{}, fmt.Errorf("failed to format json body: %w", err)
	}

	resp, err := c.conn.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&v1.TakeBookResponse{}).
		Post("/api/v1/books")
	if err != nil {
		return library.ReservedBook{}, fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return library.ReservedBook{}, fmt.Errorf("%d: %w", resp.StatusCode(), ErrInvalidStatusCode)
	}

	data, _ := resp.Result().(*v1.TakeBookResponse)

	return library.ReservedBook{
		Book:    library.Book(data.Book),
		Library: library.Info(data.Library),
	}, nil
}

func (c *Client) ReturnBook(
	_ context.Context, libraryID string, bookID string,
) (library.Book, error) {
	body, err := json.Marshal(v1.TakeBookRequest{
		BookID:    bookID,
		LibraryID: libraryID,
	})
	if err != nil {
		return library.Book{}, fmt.Errorf("failed to format json body: %w", err)
	}

	resp, err := c.conn.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("lib_id", libraryID).
		SetPathParam("book_id", bookID).
		SetBody(body).
		SetResult(&v1.ReturnBookResponse{}).
		Post("/api/v1/libraries/{lib_id}/books/{book_id}/return")
	if err != nil {
		return library.Book{}, fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return library.Book{}, fmt.Errorf("%d: %w", resp.StatusCode(), ErrInvalidStatusCode)
	}

	data, _ := resp.Result().(*v1.ReturnBookResponse)

	return library.Book(data.Book), nil
}
