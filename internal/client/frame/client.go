package frame

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	httpClient *http.Client
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

type AuthenticateReqBody struct {
	Address   string `json:"address"`
	Signature string `json:"signature"`
}

type AuthenticateRespBody struct {
	Token    string `json:"token"`
	UserInfo struct {
		Address          string  `json:"address"`
		TestnetXP        int     `json:"testnetXP"`
		HasClaimedPoints bool    `json:"hasClaimedPoints"`
		PointsClaimed    string  `json:"pointsClaimed"`
		TradesMade       int     `json:"tradesMade"`
		VolumeTraded     string  `json:"volumeTraded"`
		RoyaltiesPaid    string  `json:"royaltiesPaid"`
		TopPercent       float64 `json:"topPercent"`
		Rank             int     `json:"rank"`
		TotalAllocation  int     `json:"totalAllocation"`
	} `json:"userInfo"`
}

func (c *Client) Authenticate(reqBody *AuthenticateReqBody) (*AuthenticateRespBody, error) {
	if reqBody == nil {
		return nil, errors.New("reqBody is nil")
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marsal reqBody json: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://claim.frame-api.xyz/authenticate",
		bytes.NewReader(reqBodyJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("invalid status code: %s", resp.Status))
	}

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	defer resp.Body.Close()

	respBody := new(AuthenticateRespBody)
	if err := json.Unmarshal(respBodyBytes, respBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return respBody, nil
}

type ClaimRespBody struct {
	Message string `json:"message"`
}

func (c *Client) Claim(bearerToken string) (*ClaimRespBody, error) {
	if bearerToken == "" {
		return nil, errors.New("bearerToken is empty")
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://claim.frame-api.xyz/user/claim",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bearerToken))
	req.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(fmt.Sprintf("invalid status code: %s", resp.Status))
	}

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	defer resp.Body.Close()

	respBody := new(ClaimRespBody)
	if err := json.Unmarshal(respBodyBytes, respBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal respBody json: %w", err)
	}

	return respBody, nil
}

type UserRespBody struct {
	Address          string  `json:"address"`
	TestnetXP        int     `json:"testnetXP"`
	HasClaimedPoints bool    `json:"hasClaimedPoints"`
	PointsClaimed    string  `json:"pointsClaimed"`
	TradesMade       int     `json:"tradesMade"`
	VolumeTraded     string  `json:"volumeTraded"`
	RoyaltiesPaid    string  `json:"royaltiesPaid"`
	TopPercent       float64 `json:"topPercent"`
	Rank             int     `json:"rank"`
	TotalAllocation  int     `json:"totalAllocation"`
}

func (c *Client) User(bearerToken string) (*UserRespBody, error) {
	req, err := http.NewRequest(http.MethodGet, "https://claim.frame-api.xyz/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bearerToken))
	req.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("invalid status code: %s", resp.Status))
	}

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	respBody := new(UserRespBody)
	if err := json.Unmarshal(respBodyBytes, respBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal respBody json: %w", err)
	}

	return respBody, nil
}
