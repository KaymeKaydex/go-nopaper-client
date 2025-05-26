package go_nopaper_client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// SignatureType — type of nopaper signature.
type SignatureType string

func (s SignatureType) String() string {
	return string(s)
}

const (
	SignatureTypeServer SignatureType = "pc-server" // organization signature.
	SignatureTypeSMS    SignatureType = "pc-sms"    // sms signature for client.
)

type CreateSignatureRequest struct {
	UserGUID uuid.UUID `json:"userGuid"`
	// ResponsiblePartyForAcceptanceAct - where act will be generated. 1 - nopaper, 2 - on client side.
	// We are always use on client side.
	ResponsiblePartyForAcceptanceAct int           `json:"responsiblePartyForAcceptanceAct"`
	SignatureType                    SignatureType `json:"-"`
}

type CreateSignatureResponse struct {
	CertificateID uuid.UUID `json:"certificateId"`
}

// CreateSignature - creates new signature for client.
func (c *Client) CreateSignature(ctx context.Context, rawReq CreateSignatureRequest) (uuid.UUID, error) {
	if rawReq.UserGUID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("user uuid cant be nil")
	}
	if rawReq.ResponsiblePartyForAcceptanceAct != 1 && rawReq.ResponsiblePartyForAcceptanceAct != 2 {
		return uuid.Nil, fmt.Errorf("invalid value for ResponsiblePartyForAcceptanceAct, must be 1 or 2")
	}

	t := rawReq.SignatureType

	v := url.Values{}

	v.Add("userGuid", rawReq.UserGUID.String())
	v.Add("responsiblePartyForAcceptanceAct", strconv.Itoa(rawReq.ResponsiblePartyForAcceptanceAct))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+fmt.Sprintf("/certificate/pay-control/%s", t.String()), nil)
	if err != nil {
		return uuid.Nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	req.URL.RawQuery = v.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return uuid.Nil, err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		// Good response.
		rawResp := &CreateSignatureResponse{}

		err = json.NewDecoder(resp.Body).Decode(rawResp)
		if err != nil {
			return uuid.Nil, fmt.Errorf("cant decode good response with error: %w", err)
		}

		return rawResp.CertificateID, nil
	} else {
		bts, err := io.ReadAll(resp.Body)
		if err != nil {
			return uuid.Nil, err
		}

		return uuid.Nil, fmt.Errorf("unknown status code from nopaper: %s with status: %s", resp.Status, string(bts))
	}
}

// UserSignaturesListResponse - typed response for signature list method.
type UserSignaturesListResponse struct {
	CertificatePCServerInfoList []CertificateInfo `json:"certificateInfoList"`
}

type CertificateInfo struct {
	ID uuid.UUID `json:"certificateId"`
	// Status info.
	// status=1 (Template) — signature template.
	// status=2 (Initialization) — signature initialization.
	// status=3 (InitializationError) — signature initialization error.
	// status=4 (Available) — signature is active.
	// status=5 (Blocked) — signature is blocked.
	// status=6 (Revoked) — signature is revoked.
	Status                int       `json:"status"`
	IssuedDateTimeUtc     time.Time `json:"issuedDateTimeUtc"`
	ValidUntilDateTimeUtc time.Time `json:"validUntilDateTimeUtc"`
	OwnerName             string    `json:"ownerName"`
	// OwnerID is id of signature(certificate) owner.
	OwnerID      uuid.UUID             `json:"ownerGuid"`
	CustomData   CertificateCustomData `json:"customData"`
	ProviderType int                   `json:"providerType"`
}

type CertificateCustomData struct {
	PCUserId    string `json:"PCUserId"`
	SystemId    string `json:"SystemId"`
	PublicKey   string `json:"PublicKey"`
	IssuingType int    `json:"IssuingType"`
}

// UserSignaturesList - method that returns list of signatures for user.
func (c *Client) UserSignaturesList(ctx context.Context, userID uuid.UUID) ([]CertificateInfo, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user uuid cant be nil")
	}

	v := url.Values{}

	v.Add("userGuid", userID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url+fmt.Sprintf("/certificate/list"), nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = v.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		// Good response.
		rawResp := &UserSignaturesListResponse{}

		err = json.NewDecoder(resp.Body).Decode(rawResp)
		if err != nil {
			return nil, fmt.Errorf("cant decode good response with error: %w", err)
		}

		return rawResp.CertificatePCServerInfoList, nil
	} else {
		bts, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("unknown status code from nopaper: %s with status: %s", resp.Status, string(bts))
	}
}

// ActivateSignature activates signature for user by certificate(signature) ID.
func (c *Client) ActivateSignature(ctx context.Context, certificateID uuid.UUID) error {
	if certificateID == uuid.Nil {
		return fmt.Errorf("certificate uuid cant be nil")
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPatch, c.url+fmt.Sprintf("/certificate/pay-control/%s/activate", certificateID.String()),
		nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	return emptyResponseResult(resp)
}
