package routes

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"math"
)

func (fes *APIServer) GetBase64Nonce(ww http.ResponseWriter, req *http.Request) {
	l := 20;
	buff := make([]byte, int(math.Ceil(float64(l)/float64(1.33333333333))))
    rand.Read(buff)
    str := base64.RawURLEncoding.EncodeToString(buff)
    response := str[:l] // strip 1 extra character we get from odd length

	if err = json.NewEncoder(ww).Encode(response); err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetBase64Nonce: Problem serializing object to JSON: %v", err))
		return
	}
}