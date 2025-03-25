package hadlers

import (
	"io"
	"net/http"
)

func SendAuthorizedRequest(url, endpoint, token string) (string, error) {
	req, err := http.NewRequest("GET", url+endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
