package jsruntime

import (
	"github.com/goccy/go-spidermonkey"
	"github.com/google/uuid"
)

// defineCrypto defines the crypto object on the JS instance's global object.
func defineCrypto(js *spidermonkey.JS) error {
	cryptoObj, err := js.NewObject()
	if err != nil {
		return err
	}
	err = cryptoObj.DefineFunc("randomUUID",
		func(cfg spidermonkey.Config, args []spidermonkey.Value) (spidermonkey.Value, error) {
			// UUID v4, same as the Web standard crypto.randomUUID().
			u, err := uuid.NewRandom()
			if err != nil {
				return nil, err
			}
			return spidermonkey.ValueOf(u.String()), nil
		})
	if err != nil {
		return err
	}
	return js.Global().Set("crypto", cryptoObj)
}
