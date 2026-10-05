package mainadmin

import (
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetStatsHandler(mainAdminService *service.MainAdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := mainAdminService.GetStats()
		if err != nil {
			apperror.AbortWithException(c, apperror.DatabaseError, err)
			return
		}
		response.Success(c, stats)
	}
}
