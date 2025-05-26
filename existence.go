package go_nopaper_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserInfo struct {
	Name                string        `json:"name,omitempty"`
	Surname             string        `json:"surname,omitempty"`
	Patronymic          string        `json:"patronymic,omitempty"`
	IsShortTimePassword bool          `json:"isShortTimePassword"`
	BirthDate           time.Time     `json:"birthDate,omitempty"`
	Gender              int           `json:"gender,omitempty"`
	PassportData        *PassportData `json:"passportData,omitempty"`
}

type PassportData struct {
	Series               string    `json:"series,omitempty"`
	Number               string    `json:"number,omitempty"`
	IssuedBy             string    `json:"issuedBy,omitempty"`
	IssuingDate          time.Time `json:"issuingDate,omitempty"`
	IssuerDepartmentCode string    `json:"issuerDepartmentCode,omitempty"`
	BirthPlace           string    `json:"birthPlace,omitempty"`
}

type UserGUIDResponse struct {
	UserGUID uuid.UUID `json:"userGuid"`
}

type GetUserUUIDByPhoneErrorResponse struct {
	Code string `json:"code"`
}

// GetUserUUIDByPhone checks user existence in Nopaper system and returns user id if user exists.
func (c *Client) GetUserUUIDByPhone(ctx context.Context, phone string) (uuid.UUID, error) {
	v := url.Values{}

	v.Add("userPhone", phone)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url+"/profile-fl/user-guid/by-phone", nil)
	if err != nil {
		return uuid.Nil, err
	}

	req.URL.RawQuery = v.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return uuid.Nil, err
	}

	if resp.StatusCode == http.StatusOK {
		// Good response.
		rawResp := &UserGUIDResponse{}

		err = json.NewDecoder(resp.Body).Decode(rawResp)
		if err != nil {
			return uuid.Nil, fmt.Errorf("cant decode good response with error: %w", err)
		}

		return rawResp.UserGUID, nil
	} else if resp.StatusCode == http.StatusBadRequest {
		// Bad response.
		rawResp := &GetUserUUIDByPhoneErrorResponse{}

		err = json.NewDecoder(resp.Body).Decode(rawResp)
		if err != nil {
			return uuid.Nil, fmt.Errorf("cant decode bad response with error: %w", err)
		}

		return uuid.Nil, errorByCode(rawResp.Code)
	} else {
		return uuid.Nil, fmt.Errorf("unknown status code from nopaper: %s", resp.Status)
	}
}

type RegisterUserRequest struct {
	UserPhone string `json:"userPhone"`
	Email     string `json:"email,omitempty"`
	UserInfo
}

func (c *Client) RegisterUser(ctx context.Context, rawReq RegisterUserRequest) (uuid.UUID, error) {
	if !strings.HasPrefix(rawReq.UserPhone, "7") {
		return uuid.Nil, fmt.Errorf("user phone must be started from 7")
	}

	body := bytes.NewBuffer(nil)
	err := json.NewEncoder(body).Encode(rawReq)
	if err != nil {
		return uuid.Nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/profile-fl", body)
	if err != nil {
		return uuid.Nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return uuid.Nil, err
	}

	if resp.StatusCode == http.StatusOK {
		// Good response.
		rawResp := &UserGUIDResponse{}

		err = json.NewDecoder(resp.Body).Decode(rawResp)
		if err != nil {
			return uuid.Nil, fmt.Errorf("cant decode good response with error: %w", err)
		}

		return rawResp.UserGUID, nil
	} else {
		return uuid.Nil, fmt.Errorf("unknown status code from nopaper: %s", resp.Status)
	}
}

type PatchUserInfoRequest struct {
	UserGUID uuid.UUID `json:"userGuid"`
	UserInfo
}

// PatchUserInfo - patches current user information
func (c *Client) PatchUserInfo(ctx context.Context, rawReq PatchUserInfoRequest) error {
	if rawReq.UserGUID == uuid.Nil {
		return fmt.Errorf("user uuid cant be nil")
	}

	body := bytes.NewBuffer(nil)
	err := json.NewEncoder(body).Encode(rawReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url+"/profile-fl", body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	return emptyResponseResult(resp)
}

// EmployUser makes user an employer of company.
func (c *Client) EmployUser(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user uuid cant be nil")
	}

	body := bytes.NewBuffer(nil)
	err := json.NewEncoder(body).Encode(map[string]string{
		"userGuid": userID.String(),
	})

	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/hub/employee", body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	return emptyResponseResult(resp)
}

// FireUser makes user not an employer of a company.
func (c *Client) FireUser(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user uuid cant be nil")
	}

	v := url.Values{}

	v.Add("userGuid", userID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url+"/hub/employee", nil)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	req.URL.RawQuery = v.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	return emptyResponseResult(resp)
}
