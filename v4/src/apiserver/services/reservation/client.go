package reservation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	"ds-lab2-bmstu/pkg/readiness"
	"ds-lab2-bmstu/pkg/readiness/httpprober"
	v1 "ds-lab2-bmstu/reservation/api/http/v1"

	"ds-lab2-bmstu/apiserver/core/ports/reservation"
)

const probeKey = "http-reservation-client"

var ErrInvalidStatusCode = errors.New("invalid status code")

type Client struct {
	lg *slog.Logger

	conn *resty.Client
}

func New(lg *slog.Logger, cfg reservation.Config, probe *readiness.Probe) (*Client, error) {
	client := resty.New().
		SetTransport(&http.Transport{
			MaxIdleConns:       10,               //nolint: gomnd
			IdleConnTimeout:    30 * time.Second, //nolint: gomnd
			DisableCompression: true,
		}).
		SetBaseURL(fmt.Sprintf("http://%s", net.JoinHostPort(cfg.Host, cfg.Port)))

	c := Client{
		lg:   lg,
		conn: client,
	}

	go httpprober.New(lg, client).Ping(probeKey, probe)

	return &c, nil
}

func (c *Client) GetUserReservations(
	_ context.Context, username, status string,
) ([]reservation.Info, error) {
	q := map[string]string{}
	if status != "" {
		q["status"] = status
	}

	resp, err := c.conn.R().
		SetHeader("X-User-Name", username).
		SetQueryParams(q).
		SetResult(&[]v1.Reservation{}).
		Get("/api/v1/reservations")
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("%d: %w", resp.StatusCode(), ErrInvalidStatusCode)
	}

	data, _ := resp.Result().(*[]v1.Reservation)

	var reservs []reservation.Info
	for _, res := range *data {
		reservs = append(reservs, reservation.Info{
			ID:        res.ID,
			Username:  username,
			Status:    res.Status,
			Start:     res.Start,
			End:       res.End,
			LibraryID: res.LibraryID,
			BookID:    res.BookID,
		})
	}

	return reservs, nil
}

func (c *Client) AddUserReservation(_ context.Context, rsrvtn reservation.Info) (string, error) {
	body, err := json.Marshal(v1.AddReservationRequest{
		Status:    rsrvtn.Status,
		Start:     rsrvtn.Start,
		End:       rsrvtn.End,
		BookID:    rsrvtn.BookID,
		LibraryID: rsrvtn.LibraryID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to format json body: %w", err)
	}

	resp, err := c.conn.R().
		SetHeader("X-User-Name", rsrvtn.Username).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&v1.AddReservationResponse{}).
		Post("/api/v1/reservations")
	if err != nil {
		return "", fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("%d: %w", resp.StatusCode(), ErrInvalidStatusCode)
	}

	data, _ := resp.Result().(*v1.AddReservationResponse)

	return data.ID, nil
}

func (c *Client) SetUserReservationStatus(
	_ context.Context, id, status string,
) error {
	resp, err := c.conn.R().
		SetPathParam("id", id).
		SetQueryParam("status", status).
		Patch("/api/v1/reservations/{id}")
	if err != nil {
		return fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%d: %w", resp.StatusCode(), ErrInvalidStatusCode)
	}

	return nil
}
