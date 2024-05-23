package middleware

import (
	"bytes"
	"crypto-trade-client/common/web"
	"crypto-trade-client/common/web/sign"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-hclog"
	"io"
)

func CheckSign(keys *sign.Keys) gin.HandlerFunc {
	signKeys := keys
	return func(c *gin.Context) {
		uri := c.Request.URL.RequestURI()
		method := c.Request.Method
		wallets := c.Request.Header["X-Chain-Wallet"]
		appIds := c.Request.Header["X-Chain-Appid"]
		contentTypes := c.Request.Header["Content-Type"]
		contentType := ""
		if len(appIds) == 0 ||
			(method != "GET" && len(contentTypes) == 0) {
			hclog.L().Error("signature field missing.",
				"wallets", wallets, "appIds", appIds, "contentTypes", contentTypes)
			c.IndentedJSON(400, web.NewErrorResponse(web.ErrInvalidParam))
			c.Abort()
			return
		}
		wallet := ""
		appId := appIds[0]
		// contentType not exist when request method is GET
		if len(contentTypes) > 0 {
			contentType = contentTypes[0]
		}

		var bodyBytes []byte
		var err error
		defer c.Request.Body.Close()
		if bodyBytes, err = io.ReadAll(c.Request.Body); err != nil {
			c.IndentedJSON(500, web.NewErrorResponse(web.ErrInternal))
			c.Abort()
			return
		}
		bs := sign.Construct(uri, method, contentType, wallet, appId, bodyBytes)

		sigs := c.Request.Header["X-Chain-Sign"]
		if sigs == nil || len(sigs) == 0 {
			hclog.L().Error("signature field missing.", "sign", sigs)
			c.IndentedJSON(400, web.NewErrorResponse(web.ErrInvalidParam))
			c.Abort()
			return
		}
		if !signKeys.Verify(bs, sigs[0], wallet, appId, sign.Secp256k1) {
			c.IndentedJSON(401, web.NewErrorResponse(web.ErrResourceUnauthorized))
			c.Abort()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Set("walletName", wallet)
		c.Set("appId", appId)
		c.Next()
	}
}
