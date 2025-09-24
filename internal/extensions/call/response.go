package call

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kyma-project/cli.v3/internal/clierror"
)

type errorResponse struct {
	Error string `json:"error"`
}

func decodeResponse(resp *http.Response) ([]byte, clierror.Error) {
	bytesBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to read response body"))
	}

	// handle error responses
	if resp.StatusCode >= 400 {
		var errResp errorResponse
		if err := json.Unmarshal(bytesBody, &errResp); err != nil {
			// return generic error message if response cannot be decoded
			return nil, clierror.New(fmt.Sprintf("request failed with status code %d: %s", resp.StatusCode, bytesBody))
		}

		return nil, clierror.New(errResp.Error)
	}

	return bytesBody, nil
}

type FileResponse struct {
	Name string `json:"name"`
	Data string `json:"data"` // base64 encoded file content
}

type FilesListResponse struct {
	OutputMessage string         `json:"outputMessage"`
	Files         []FileResponse `json:"files"`
}
