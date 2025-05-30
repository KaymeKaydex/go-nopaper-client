package nopaper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

// DocumentRouteType represents the route type of document.
type DocumentRouteType int

const (
	// Consistent is sequential mode of document route.
	// It means that order of signatures is important.
	// First signature is a signature of first person in draft document recipient sequence.
	Consistent DocumentRouteType = 1
	// Parallel is parallel mode of document route.
	// It means that order of signatures is not important.
	// All members of document chain can sign it parallel.
	Parallel DocumentRouteType = 2
)

// CreateDraftDocumentRequest is request for document(document for Nopaper is a chain of files) creation.
type CreateDraftDocumentRequest struct {
	// Title is name of your document chain in Nopaper UI, nothing else.
	Title string `json:"title,omitempty"`
	// UserID is an identifier of user that creates new draft document.
	// This might be empty.
	UserID *uuid.UUID `json:"userGuid,omitempty"`
	// RecipientInfoList is required.
	// This field describes participants of deal.
	RecipientInfoList []RecipientInfo `json:"recipientInfoList,omitempty"`
	// DocumentRouteType - type of document sign route.
	DocumentRouteType DocumentRouteType `json:"documentRouteType,omitempty"`
	// DisableChange - Setting - editing document route
	// False - each document participant can edit the route.
	// True - editing is prohibited, with the exception of:
	// 1) If the document owner is an individual, then the following can edit the route:
	// the document owner;
	// 2) If the document owner is a legal entity, then the following can edit the route:
	// the employee who is the document owner;
	// the employee of the company who is the document owner with the right.
	DisableChange bool `json:"isDisableChange"`
}

type RecipientInfo struct {
	UserPhone  string `json:"userPhone,omitempty"`
	CompanyInn string `json:"companyInn,omitempty"`
	ActionType int    `json:"actionType,omitempty"`
	SignType   int    `json:"signType,omitempty"`
	CompanyKpp string `json:"companyKpp,omitempty"`
}

type CreateDraftDocumentResponse struct {
	DocumentID int `json:"documentId"`
}

// CreateDraftDocument - creates new draft document package.
// Document for Nopaper is a chain of word or pdf files.
func (c *Client) CreateDraftDocument(ctx context.Context, rawReq CreateDraftDocumentRequest) (int, error) {
	body := bytes.NewBuffer(nil)
	err := json.NewEncoder(body).Encode(rawReq)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/document/draft", body)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		// Good response.
		rawResp := &CreateDraftDocumentResponse{}

		err = json.NewDecoder(resp.Body).Decode(rawResp)
		if err != nil {
			return 0, fmt.Errorf("cant decode good response with error: %w", err)
		}

		return rawResp.DocumentID, nil
	} else {
		bts, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}

		return 0, fmt.Errorf("unknown status code from nopaper: %s with status: %s", resp.Status, string(bts))
	}
}

type AttachFile2DocumentRequest struct {
	FileInfo FileInfo `json:"fileInfo"`
}

type FileInfo struct {
	FileNameWithExtension string `json:"fileNameWithExtension"`
	Filebase64            string `json:"filebase64"`
}

func (c *Client) AttachFile2Document(ctx context.Context, documentID int, rawReq AttachFile2DocumentRequest) error {
	if documentID == 0 {
		return fmt.Errorf("document id is required")
	}

	if rawReq.FileInfo.FileNameWithExtension == "" {
		return fmt.Errorf("filename with extension can not be empty")
	}

	if rawReq.FileInfo.Filebase64 == "" {
		return fmt.Errorf("filebase64 can not be empty")
	}

	body := bytes.NewBuffer(nil)
	err := json.NewEncoder(body).Encode(rawReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+fmt.Sprintf("/document/%d/file", documentID), body)
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

// ActivateDocument - activates document and changes status to active state.
func (c *Client) ActivateDocument(ctx context.Context, documentID int) error {
	if documentID == 0 {
		return fmt.Errorf("document id can not be empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+fmt.Sprintf("/document/%d/send", documentID), nil)
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

// StartSMSSignatureProcess sends sms for user recipient of deal.
// This method starts process of sign, then user must send SMS code to system, and we approve his signature.
func (c *Client) StartSMSSignatureProcess(ctx context.Context, documentID int, signatureID uuid.UUID) error {
	if signatureID == uuid.Nil {
		return fmt.Errorf("signature id can not be empty")
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		c.url+fmt.Sprintf("/document/%d/sign/pc-sms/%s", documentID, signatureID.String()),
		nil)
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

func (c *Client) SignViaServerSignature(ctx context.Context, documentID int, signatureID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPut,
		c.url+fmt.Sprintf("/document/%d/sign/pc-server/%s", documentID, signatureID.String()),
		nil)
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

func (c *Client) ConfirmSMSSign(ctx context.Context, documentID int, signatureID uuid.UUID, code string) error {
	b := bytes.NewBuffer(nil)
	err := json.NewEncoder(b).Encode(map[string]string{
		"code": code,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		c.url+fmt.Sprintf("/document/%d/sign/pc-sms/%s/confirm", documentID, signatureID.String()),
		b)
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

type GetFileIDsInDocumentResponse struct {
	OriginFileList           []FileIDInfo `json:"originFileList"`
	OriginFileWithStampList  []FileIDInfo `json:"originFileWithStampList"`
	OfertaList               []FileIDInfo `json:"ofertaList"`
	OfertaWithStampList      []FileIDInfo `json:"ofertaWithStampList"`
	ProcuratoryList          []FileIDInfo `json:"procuratoryList"`
	ProcuratoryWithStampList []FileIDInfo `json:"procuratoryWithStampList"`
}

type FileIDInfo struct {
	FileID                  string    `json:"fileId"`
	OriginNameWithExtension string    `json:"originNameWithExtension"`
	SizeKb                  int       `json:"sizeKb"`
	OriginalFileId          uuid.UUID `json:"originalFileId"`
}

func (c *Client) GetFileIDsInDocument(ctx context.Context, documentID int) (*GetFileIDsInDocumentResponse, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		c.url+fmt.Sprintf("/document/%d/file-info/list", documentID),
		nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		// Good response.
		rawResp := &GetFileIDsInDocumentResponse{}

		err = json.NewDecoder(resp.Body).Decode(rawResp)
		if err != nil {
			return nil, fmt.Errorf("cant decode good response with error: %w", err)
		}

		return rawResp, nil
	} else {
		bts, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("unknown status code from nopaper: %s with status: %s", resp.Status, string(bts))
	}
}

type GetFilesByIDRequest struct {
	FileID     string `json:"fileId"`
	DocumentID int    `json:"documentId"`
}

type FileInfoResponse struct {
	FileID                uuid.UUID `json:"fileId"`
	FileBase64            string    `json:"fileBase64"`
	FileNameWithExtension string    `json:"fileNameWithExtension"`
}

type GetFilesByIDResponse struct {
	FileInfoList []FileInfoResponse `json:"fileInfoList"`
}

func (c *Client) GetFilesByID(ctx context.Context, rawReq []GetFilesByIDRequest) ([]FileInfoResponse, error) {
	body := bytes.NewBuffer(nil)
	err := json.NewEncoder(body).Encode(map[string][]GetFilesByIDRequest{
		"documentFileInfoList": rawReq,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		c.url+fmt.Sprintf("/document/file/list"),
		body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		// Good response.
		rawResp := &GetFilesByIDResponse{}

		err = json.NewDecoder(resp.Body).Decode(&rawResp)
		if err != nil {
			return nil, fmt.Errorf("cant decode good response with error: %w", err)
		}

		return rawResp.FileInfoList, nil
	} else {
		bts, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("unknown status code from nopaper: %s with status: %s", resp.Status, string(bts))
	}
}
