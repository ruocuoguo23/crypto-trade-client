package middleware

import (
	"crypto-trade-client/common/web"
	"github.com/gin-gonic/gin"
)

func JSONAppErrorReporter() gin.HandlerFunc {
	return jsonAppErrorReporterT(gin.ErrorTypeAny)
}

func jsonAppErrorReporterT(errType gin.ErrorType) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		detectedErrors := c.Errors.ByType(errType)

		if len(detectedErrors) > 0 {
			err := detectedErrors[0].Err
			var parsedError web.Error
			switch err.(type) {
			case web.Error:
				parsedError = err.(web.Error)
			default:
				parsedError = web.ErrInternal
			}

			hclog.L().Named("gin-error").Warn(parsedError.Error())

			// Put the error into response
			c.IndentedJSON(parsedError.Code(), web.NewErrorResponse(parsedError))
			c.Abort()
			// or c.AbortWithStatusJSON(parsedError.Code, parsedError)
			return
		}
	}
}
